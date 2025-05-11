package queues

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// MQTTPayload represents the payload format received from ESP devices
type MQTTPayload struct {
	APIKEY     string `json:"X-API-KEY"`
	MacAddress string `json:"X-MAC-ADDRESS"`
	Image      string `json:"image"`
}

// WebSocketClient represents a client connected to the websocket
type WebSocketClient struct {
	ESPHmac  string      // MAC address of ESP this client is watching (empty for "all" clients)
	SendChan chan []byte // Channel to send data to the client
}

type ScannerMQTT struct {
	Client         mqtt.Client
	Token          mqtt.Token
	S3Service      *services.S3Service
	SubscribeTopic string
	PublishTopic   string

	// Last received image data per ESP MAC
	DeviceImages map[string]*models.ParkingImage
	ImagesLock   sync.RWMutex

	// WebSocket related fields
	SpecificClients map[string][]*WebSocketClient // Map ESP MAC to clients
	AllClients      []*WebSocketClient            // Clients watching all devices
	ClientsLock     sync.RWMutex                  // To safely access the clients map

	// Connection state
	isConnected bool
	connLock    sync.RWMutex
}

func NewScannerMQTT(s3Service *services.S3Service) *ScannerMQTT {
	scanner := &ScannerMQTT{
		S3Service:       s3Service,
		SubscribeTopic:  viper.GetString("mqtt.topic.subscribe"),
		PublishTopic:    viper.GetString("mqtt.topic.publish"),
		DeviceImages:    make(map[string]*models.ParkingImage),
		SpecificClients: make(map[string][]*WebSocketClient),
		AllClients:      make([]*WebSocketClient, 0),
		isConnected:     false,
	}

	// Connect to MQTT broker
	scanner.connectMQTT()

	// Start debug ticker
	scanner.StartDebugTicker()

	return scanner
}

func (s *ScannerMQTT) connectMQTT() {
	// Configure MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", viper.GetString("mqtt.host"), viper.GetInt("mqtt.port")))

	// Add timestamp to avoid client ID conflicts
	clientID := fmt.Sprintf("parkingo-core-%d", time.Now().Unix())
	opts.SetClientID(clientID)
	logrus.Infof("Connecting to MQTT with client ID: %s", clientID)

	// Configure automatic reconnect
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(1 * time.Minute)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetCleanSession(false) // Keep subscriptions on reconnect

	// Set connection timeout
	opts.SetConnectTimeout(10 * time.Second)

	// Configure keep alive
	opts.SetKeepAlive(60 * time.Second)

	// Set up connection handlers
	opts.OnConnect = func(c mqtt.Client) {
		s.connLock.Lock()
		s.isConnected = true
		s.connLock.Unlock()

		logrus.Info("Connected to MQTT broker")

		// Subscribe to topic on connect/reconnect
		s.subscribeToTopic()
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		s.connLock.Lock()
		s.isConnected = false
		s.connLock.Unlock()

		logrus.Errorf("Connection lost to MQTT broker: %v, will auto-reconnect", err)
	}

	opts.OnReconnecting = func(c mqtt.Client, opts *mqtt.ClientOptions) {
		logrus.Info("Attempting to reconnect to MQTT broker...")
	}

	// Create and connect client
	client := mqtt.NewClient(opts)
	token := client.Connect()
	waitTimeout := 10 * time.Second
	if !token.WaitTimeout(waitTimeout) {
		logrus.Warnf("MQTT connection timed out after %s, retrying in background", waitTimeout)
	} else if err := token.Error(); err != nil {
		logrus.Errorf("MQTT connection error: %v, will retry in background", err)
	}

	s.Client = client
	s.Token = token
}

func (s *ScannerMQTT) isConnectedToMQTT() bool {
	s.connLock.RLock()
	defer s.connLock.RUnlock()
	return s.isConnected && s.Client.IsConnected()
}

func (s *ScannerMQTT) subscribeToTopic() {
	if !s.Client.IsConnected() {
		logrus.Warn("Cannot subscribe: MQTT client is not connected")
		return
	}

	logrus.Infof("Subscribing to MQTT topic: %s", s.SubscribeTopic)
	token := s.Client.Subscribe(s.SubscribeTopic, 1, s.handleMQTTMessage)
	if !token.WaitTimeout(5 * time.Second) {
		logrus.Warnf("Subscription to topic %s timed out", s.SubscribeTopic)
		return
	}

	if err := token.Error(); err != nil {
		logrus.Errorf("Failed to subscribe to topic %s: %v", s.SubscribeTopic, err)
		return
	}

	logrus.Infof("Successfully subscribed to topic: %s", s.SubscribeTopic)
}

func (s *ScannerMQTT) Subscribe() {
	// Initial subscription will be handled in OnConnect handler
	// Subsequent calls will check if we need to resubscribe
	if !s.isConnectedToMQTT() {
		logrus.Warn("MQTT client is not connected, attempting to reconnect...")
		s.Client.Connect()
		return
	}

	// Check if we're already subscribed
	s.subscribeToTopic()
}

func (s *ScannerMQTT) handleMQTTMessage(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	logrus.Infof("Received message from topic: %s", topic)
	logrus.Debugf("Raw MQTT payload: %s", payload)

	// Parse the payload
	var mqttPayload MQTTPayload
	if err := json.Unmarshal(msg.Payload(), &mqttPayload); err != nil {
		logrus.Errorf("Failed to unmarshal MQTT payload: %v. Raw payload: %s", err, payload)
		return
	}

	logrus.Infof("Successfully parsed MQTT payload. MAC Address: %s, API Key: %s, Image length: %d bytes",
		mqttPayload.MacAddress, mqttPayload.APIKEY, len(mqttPayload.Image))

	// Verify API key if needed
	expectedAPIKey := viper.GetString("mqtt.api_key")
	if expectedAPIKey != "" && mqttPayload.APIKEY != expectedAPIKey {
		logrus.Warnf("Invalid API key in MQTT message: %s", mqttPayload.APIKEY)
		return
	}

	// Skip empty MAC addresses
	if mqttPayload.MacAddress == "" {
		logrus.Warn("Received MQTT message with empty MAC address, ignoring")
		return
	}

	// Add image data format prefix if not already present
	imageData := mqttPayload.Image
	if len(imageData) > 0 && !isBase64ImagePrefix(imageData) {
		imageData = "data:image/jpeg;base64," + imageData
		logrus.Debug("Added base64 image prefix to raw image data")
	}

	// Calculate an image hash to check for duplicate images
	imageHash := ""
	if len(imageData) > 100 {
		// Use the first 100 chars as a simple hash
		imageHash = imageData[:100]
	} else {
		imageHash = imageData
	}

	// Check if this is a duplicate of the most recent image
	s.ImagesLock.RLock()
	previousImage, exists := s.DeviceImages[mqttPayload.MacAddress]
	s.ImagesLock.RUnlock()

	currentTime := time.Now().UnixMilli()

	if exists && previousImage != nil {
		// Calculate previous hash
		prevHash := ""
		if len(previousImage.ImageData) > 100 {
			prevHash = previousImage.ImageData[:100]
		} else {
			prevHash = previousImage.ImageData
		}

		if prevHash == imageHash {
			logrus.Warnf("Received duplicate image from ESP MAC: %s, ignoring or updating timestamp only",
				mqttPayload.MacAddress)

			// Still update timestamp to show it's active
			s.ImagesLock.Lock()
			previousImage.Timestamp = currentTime
			s.ImagesLock.Unlock()

			// Broadcast with updated timestamp (image data is the same)
			s.broadcastToClients(previousImage)
			return
		}
	}

	// Create the parking image data with new timestamp
	parkingImage := &models.ParkingImage{
		ESPHmac:   mqttPayload.MacAddress,
		ImageData: imageData,
		Timestamp: currentTime,
	}

	// Store the latest image for this ESP
	s.ImagesLock.Lock()
	s.DeviceImages[mqttPayload.MacAddress] = parkingImage
	deviceCount := len(s.DeviceImages)
	s.ImagesLock.Unlock()

	logrus.Infof("Stored NEW image for ESP MAC: %s with timestamp %d. Total devices in memory: %d",
		mqttPayload.MacAddress, currentTime, deviceCount)

	// Check if we actually have any WebSocket clients before attempting broadcast
	s.ClientsLock.RLock()
	specificClientsCount := 0
	if clients, ok := s.SpecificClients[mqttPayload.MacAddress]; ok {
		specificClientsCount = len(clients)
	}
	allClientsCount := len(s.AllClients)
	s.ClientsLock.RUnlock()

	if specificClientsCount == 0 && allClientsCount == 0 {
		logrus.Warn("No WebSocket clients registered, not broadcasting this message")
		// Force a status update to see what's going on
		s.PrintConnectionStatus()
		return
	}

	// Broadcast to all relevant WebSocket clients
	s.broadcastToClients(parkingImage)
}

// isBase64ImagePrefix checks if the string already has a data URL prefix
func isBase64ImagePrefix(data string) bool {
	return len(data) > 11 && (data[:11] == "data:image/" || data[:11] == "data:video/")
}

// GetAllDevices returns the list of all ESP devices that have sent images
func (s *ScannerMQTT) GetAllDevices() []*models.ParkingImage {
	s.ImagesLock.RLock()
	defer s.ImagesLock.RUnlock()

	devices := make([]*models.ParkingImage, 0, len(s.DeviceImages))
	for _, image := range s.DeviceImages {
		devices = append(devices, image)
	}

	logrus.Infof("GetAllDevices called, returning %d devices", len(devices))
	return devices
}

// GetDeviceImage returns the latest image for a specific ESP device
func (s *ScannerMQTT) GetDeviceImage(espHmac string) *models.ParkingImage {
	s.ImagesLock.RLock()
	defer s.ImagesLock.RUnlock()

	image := s.DeviceImages[espHmac]
	if image == nil {
		logrus.Warnf("GetDeviceImage: No image found for ESP MAC: %s", espHmac)
	} else {
		logrus.Infof("GetDeviceImage: Found image for ESP MAC: %s", espHmac)
	}

	return image
}

// RegisterSpecificClient registers a client for a specific ESP MAC address
func (s *ScannerMQTT) RegisterSpecificClient(espHmac string, sendChan chan []byte) {
	if espHmac == "" {
		logrus.Error("RegisterSpecificClient called with empty ESP MAC")
		return
	}

	if sendChan == nil {
		logrus.Error("RegisterSpecificClient called with nil channel")
		return
	}

	client := &WebSocketClient{
		ESPHmac:  espHmac,
		SendChan: sendChan,
	}

	s.ClientsLock.Lock()
	defer s.ClientsLock.Unlock()

	// Log current client counts before registration
	specificCount := 0
	for _, clients := range s.SpecificClients {
		specificCount += len(clients)
	}
	allCount := len(s.AllClients)
	logrus.Infof("BEFORE REGISTER: Current client counts - Specific: %d, All: %d",
		specificCount, allCount)

	if _, ok := s.SpecificClients[espHmac]; !ok {
		s.SpecificClients[espHmac] = make([]*WebSocketClient, 0)
	}

	s.SpecificClients[espHmac] = append(s.SpecificClients[espHmac], client)
	logrus.Infof("New client registered for ESP MAC: %s. Total clients: %d",
		espHmac, len(s.SpecificClients[espHmac]))

	// Log current client counts after registration
	specificCount = 0
	for _, clients := range s.SpecificClients {
		specificCount += len(clients)
	}
	allCount = len(s.AllClients)
	logrus.Infof("AFTER REGISTER: Updated client counts - Specific: %d, All: %d",
		specificCount, allCount)

	// Send current image if available
	s.ImagesLock.RLock()
	deviceImage, exists := s.DeviceImages[espHmac]
	s.ImagesLock.RUnlock()

	if exists {
		logrus.Infof("Sending existing image data to new client for ESP MAC: %s", espHmac)
		data, err := json.Marshal(deviceImage)
		if err == nil {
			select {
			case sendChan <- data:
				logrus.Infof("Successfully sent existing image to new client for ESP MAC: %s", espHmac)
			default:
				logrus.Warnf("Failed to send existing image to new client for ESP MAC: %s (channel full or closed)", espHmac)
			}
		} else {
			logrus.Errorf("Failed to marshal existing image for new client: %v", err)
		}
	} else {
		logrus.Infof("No existing image data available for ESP MAC: %s", espHmac)
	}
}

// RegisterAllClient registers a client to receive updates for all ESP devices
func (s *ScannerMQTT) RegisterAllClient(sendChan chan []byte) {
	if sendChan == nil {
		logrus.Error("RegisterAllClient called with nil channel")
		return
	}

	client := &WebSocketClient{
		ESPHmac:  "", // Empty means "all devices"
		SendChan: sendChan,
	}

	s.ClientsLock.Lock()

	// Log current client counts before registration
	specificCount := 0
	for _, clients := range s.SpecificClients {
		specificCount += len(clients)
	}
	allCount := len(s.AllClients)
	logrus.Infof("BEFORE REGISTER: Current client counts - Specific: %d, All: %d",
		specificCount, allCount)

	s.AllClients = append(s.AllClients, client)
	clientCount := len(s.AllClients)

	// Log current client counts after registration
	specificCount = 0
	for _, clients := range s.SpecificClients {
		specificCount += len(clients)
	}
	allCount = len(s.AllClients)
	logrus.Infof("AFTER REGISTER: Updated client counts - Specific: %d, All: %d",
		specificCount, allCount)

	s.ClientsLock.Unlock()

	logrus.Infof("New client registered for ALL devices. Total 'all' clients: %d", clientCount)

	// Send all current images
	s.ImagesLock.RLock()
	allDevices := make([]*models.ParkingImage, 0, len(s.DeviceImages))
	for _, image := range s.DeviceImages {
		allDevices = append(allDevices, image)
	}
	s.ImagesLock.RUnlock()

	if len(allDevices) > 0 {
		logrus.Infof("Sending %d existing device images to new 'all devices' client", len(allDevices))
		data, err := json.Marshal(allDevices)
		if err == nil {
			select {
			case sendChan <- data:
				logrus.Info("Successfully sent existing devices data to new 'all devices' client")
			default:
				logrus.Warn("Failed to send existing devices data to new 'all devices' client (channel full or closed)")
			}
		} else {
			logrus.Errorf("Failed to marshal existing devices data for new 'all devices' client: %v", err)
		}
	} else {
		logrus.Info("No existing device images available to send to new 'all devices' client")
	}
}

// UnregisterClient removes a WebSocket client
func (s *ScannerMQTT) UnregisterClient(espHmac string, sendChan chan []byte) {
	s.ClientsLock.Lock()
	defer s.ClientsLock.Unlock()

	if espHmac == "" {
		// Remove from "all devices" clients
		for i, client := range s.AllClients {
			if client.SendChan == sendChan {
				s.AllClients = append(s.AllClients[:i], s.AllClients[i+1:]...)
				logrus.Infof("Client unregistered from ALL devices. Remaining clients: %d",
					len(s.AllClients))
				break
			}
		}
	} else {
		// Remove from specific clients
		if clients, ok := s.SpecificClients[espHmac]; ok {
			for i, client := range clients {
				if client.SendChan == sendChan {
					s.SpecificClients[espHmac] = append(clients[:i], clients[i+1:]...)
					logrus.Infof("Client unregistered for ESP MAC: %s. Remaining clients: %d",
						espHmac, len(s.SpecificClients[espHmac]))
					break
				}
			}

			// If no clients left for this ESP MAC, remove the key
			if len(s.SpecificClients[espHmac]) == 0 {
				delete(s.SpecificClients, espHmac)
				logrus.Infof("Removed empty client list for ESP MAC: %s", espHmac)
			}
		}
	}
}

// broadcastToClients sends the image data to relevant clients
func (s *ScannerMQTT) broadcastToClients(parkingImage *models.ParkingImage) {
	if parkingImage == nil {
		logrus.Error("Attempted to broadcast nil parkingImage")
		return
	}

	logrus.Infof("Starting broadcast for ESP MAC: %s, image size: %d bytes, timestamp: %d",
		parkingImage.ESPHmac, len(parkingImage.ImageData), parkingImage.Timestamp)

	s.ClientsLock.RLock()
	defer s.ClientsLock.RUnlock()

	// Log client counts for debugging
	specificClientCount := 0
	if clients, ok := s.SpecificClients[parkingImage.ESPHmac]; ok {
		specificClientCount = len(clients)
	}
	allClientCount := len(s.AllClients)

	logrus.Infof("Broadcasting image for ESP MAC %s to %d specific clients and %d 'all' clients",
		parkingImage.ESPHmac, specificClientCount, allClientCount)

	// Prepare data for specific clients
	specificData, err := json.Marshal(parkingImage)
	if err != nil {
		logrus.Errorf("Failed to marshal image data for broadcast: %v", err)
		return
	}

	// Debug: show part of the data being sent
	if len(specificData) > 50 {
		dataPreview := string(specificData[:50])
		logrus.Debugf("Data to send (truncated): %s...", dataPreview)

		// Debug the JSON structure to verify timestamp is included
		var debugObj map[string]interface{}
		if err := json.Unmarshal(specificData, &debugObj); err == nil {
			logrus.Debugf("JSON data fields: esp_hmac=%v, timestamp=%v, image_data_length=%d",
				debugObj["esp_hmac"], debugObj["timestamp"], len(fmt.Sprintf("%v", debugObj["image_data"])))
		}
	}

	// Send to clients watching this specific ESP
	if clients, ok := s.SpecificClients[parkingImage.ESPHmac]; ok && len(clients) > 0 {
		logrus.Infof("Broadcasting to %d specific clients for ESP MAC: %s",
			len(clients), parkingImage.ESPHmac)

		for i, client := range clients {
			if client == nil || client.SendChan == nil {
				logrus.Warnf("Found nil client or channel at index %d for ESP MAC: %s", i, parkingImage.ESPHmac)
				continue
			}

			select {
			case client.SendChan <- specificData:
				logrus.Debugf("Successfully sent image update to specific client %d for ESP MAC: %s",
					i, parkingImage.ESPHmac)
			default:
				logrus.Warnf("Failed to send message to client %d for ESP MAC: %s (channel full or closed)",
					i, parkingImage.ESPHmac)
			}
		}
		logrus.Infof("Sent image update to %d clients for ESP MAC: %s",
			len(clients), parkingImage.ESPHmac)
	} else {
		logrus.Infof("No specific clients registered for ESP MAC: %s", parkingImage.ESPHmac)
	}

	// If we have clients watching all devices, send to them as an array with one item
	if len(s.AllClients) > 0 {
		logrus.Infof("Broadcasting to %d 'all devices' clients", len(s.AllClients))

		// For "all" clients, we need to wrap it in an array
		allData, err := json.Marshal([]*models.ParkingImage{parkingImage})
		if err != nil {
			logrus.Errorf("Failed to marshal device array for 'all' clients: %v", err)
			return
		}

		// Debug the array JSON to verify the structure
		if len(allData) > 50 {
			logrus.Debugf("Array data to send (truncated): %s...", string(allData[:50]))

			// Debug the JSON array structure
			var debugArray []map[string]interface{}
			if err := json.Unmarshal(allData, &debugArray); err == nil && len(debugArray) > 0 {
				logrus.Debugf("JSON array item: esp_hmac=%v, timestamp=%v, image_data_length=%d",
					debugArray[0]["esp_hmac"], debugArray[0]["timestamp"],
					len(fmt.Sprintf("%v", debugArray[0]["image_data"])))
			}
		}

		successCount := 0
		for i, client := range s.AllClients {
			if client == nil || client.SendChan == nil {
				logrus.Warnf("Found nil client or channel at index %d for 'all devices'", i)
				continue
			}

			select {
			case client.SendChan <- allData:
				logrus.Debugf("Successfully sent image update to 'all devices' client %d", i)
				successCount++
			default:
				logrus.Warnf("Failed to send message to 'all devices' client %d (channel full or closed)", i)
			}
		}
		logrus.Infof("Successfully sent image update to %d/%d 'all devices' clients",
			successCount, len(s.AllClients))
	} else {
		logrus.Info("No 'all devices' clients registered")
	}

	logrus.Infof("Completed broadcasting image for ESP MAC: %s", parkingImage.ESPHmac)
}

// BroadcastTestMessage sends a test message to clients
func (s *ScannerMQTT) BroadcastTestMessage(espHmac string, isTestMessage bool) {
	// Create a test parking image
	testMessage := "TEST_IMAGE_DATA"
	if isTestMessage {
		testMessage += "_TEST"
	}

	parkingImage := &models.ParkingImage{
		ESPHmac:   espHmac,
		ImageData: "data:image/jpeg;base64," + testMessage,
		Timestamp: time.Now().UnixMilli(),
	}

	logrus.Infof("Broadcasting test message for ESP MAC: %s", espHmac)
	s.broadcastToClients(parkingImage)
}

// StartDebugTicker starts a ticker to periodically log debug information
func (s *ScannerMQTT) StartDebugTicker() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			s.PrintConnectionStatus()
		}
	}()
	logrus.Info("Started debug ticker for WebSocket client tracking")
}

// PrintConnectionStatus prints the current connection status and client counts
func (s *ScannerMQTT) PrintConnectionStatus() {
	s.ClientsLock.RLock()
	defer s.ClientsLock.RUnlock()

	// Count specific clients
	specificCount := 0
	for espMac, clients := range s.SpecificClients {
		clientCount := len(clients)
		specificCount += clientCount
		if clientCount > 0 {
			logrus.Infof("STATUS: There are %d specific clients for ESP MAC: %s", clientCount, espMac)
		}
	}

	// Count all clients
	allCount := len(s.AllClients)

	// Log total MQTT connection status
	s.connLock.RLock()
	mqttConnected := s.isConnected && s.Client.IsConnected()
	s.connLock.RUnlock()

	// Total device count
	s.ImagesLock.RLock()
	deviceCount := len(s.DeviceImages)
	s.ImagesLock.RUnlock()

	logrus.Infof("STATUS: MQTT connected: %v, Total devices: %d, Specific clients: %d, All clients: %d",
		mqttConnected, deviceCount, specificCount, allCount)
}
