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

	return scanner
}

func (s *ScannerMQTT) connectMQTT() {
	// Configure MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", viper.GetString("mqtt.host"), viper.GetInt("mqtt.port")))
	opts.SetClientID("parkingo-core-" + fmt.Sprintf("%d", time.Now().Unix())) // Add timestamp to avoid client ID conflicts

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
	logrus.Infof("Received message from topic: %s", msg.Topic())

	// Parse the payload
	var payload MQTTPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		logrus.Errorf("Failed to unmarshal MQTT payload: %v", err)
		return
	}

	// Verify API key if needed
	expectedAPIKey := viper.GetString("mqtt.api_key")
	if expectedAPIKey != "" && payload.APIKEY != expectedAPIKey {
		logrus.Warnf("Invalid API key in MQTT message: %s", payload.APIKEY)
		return
	}

	// Skip empty MAC addresses
	if payload.MacAddress == "" {
		logrus.Warn("Received MQTT message with empty MAC address, ignoring")
		return
	}

	// Add image data format prefix if not already present
	imageData := payload.Image
	if len(imageData) > 0 && !isBase64ImagePrefix(imageData) {
		imageData = "data:image/jpeg;base64," + imageData
	}

	// Create the parking image data
	parkingImage := &models.ParkingImage{
		ESPHmac:   payload.MacAddress,
		ImageData: imageData,
		Timestamp: time.Now().UnixMilli(),
	}

	// Store the latest image for this ESP
	s.ImagesLock.Lock()
	s.DeviceImages[payload.MacAddress] = parkingImage
	s.ImagesLock.Unlock()

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

	return devices
}

// GetDeviceImage returns the latest image for a specific ESP device
func (s *ScannerMQTT) GetDeviceImage(espHmac string) *models.ParkingImage {
	s.ImagesLock.RLock()
	defer s.ImagesLock.RUnlock()

	return s.DeviceImages[espHmac]
}

// RegisterSpecificClient registers a client for a specific ESP MAC address
func (s *ScannerMQTT) RegisterSpecificClient(espHmac string, sendChan chan []byte) {
	client := &WebSocketClient{
		ESPHmac:  espHmac,
		SendChan: sendChan,
	}

	s.ClientsLock.Lock()
	defer s.ClientsLock.Unlock()

	if _, ok := s.SpecificClients[espHmac]; !ok {
		s.SpecificClients[espHmac] = make([]*WebSocketClient, 0)
	}

	s.SpecificClients[espHmac] = append(s.SpecificClients[espHmac], client)
	logrus.Infof("New client registered for ESP MAC: %s. Total clients: %d",
		espHmac, len(s.SpecificClients[espHmac]))

	// Send current image if available
	s.ImagesLock.RLock()
	deviceImage, exists := s.DeviceImages[espHmac]
	s.ImagesLock.RUnlock()

	if exists {
		data, err := json.Marshal(deviceImage)
		if err == nil {
			sendChan <- data
		}
	}
}

// RegisterAllClient registers a client to receive updates for all ESP devices
func (s *ScannerMQTT) RegisterAllClient(sendChan chan []byte) {
	client := &WebSocketClient{
		ESPHmac:  "", // Empty means "all devices"
		SendChan: sendChan,
	}

	s.ClientsLock.Lock()
	s.AllClients = append(s.AllClients, client)
	clientCount := len(s.AllClients)
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
		data, err := json.Marshal(allDevices)
		if err == nil {
			sendChan <- data
		}
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
			}
		}
	}
}

// broadcastToClients sends the image data to relevant clients
func (s *ScannerMQTT) broadcastToClients(parkingImage *models.ParkingImage) {
	s.ClientsLock.RLock()
	defer s.ClientsLock.RUnlock()

	// Prepare data for specific clients
	specificData, err := json.Marshal(parkingImage)
	if err != nil {
		logrus.Errorf("Failed to marshal image data for broadcast: %v", err)
		return
	}

	// Send to clients watching this specific ESP
	if clients, ok := s.SpecificClients[parkingImage.ESPHmac]; ok {
		for _, client := range clients {
			select {
			case client.SendChan <- specificData:
				// Successfully sent
			default:
				logrus.Warnf("Failed to send message to client for ESP MAC: %s", parkingImage.ESPHmac)
			}
		}
		logrus.Infof("Sent image update to %d clients for ESP MAC: %s",
			len(clients), parkingImage.ESPHmac)
	}

	// If we have clients watching all devices, send to them as an array with one item
	if len(s.AllClients) > 0 {
		// For "all" clients, we need to wrap it in an array
		allData, err := json.Marshal([]*models.ParkingImage{parkingImage})
		if err != nil {
			logrus.Errorf("Failed to marshal image array for 'all' clients: %v", err)
			return
		}

		for _, client := range s.AllClients {
			select {
			case client.SendChan <- allData:
				// Successfully sent
			default:
				logrus.Warn("Failed to send message to 'all devices' client")
			}
		}
		logrus.Infof("Sent image update to %d 'all devices' clients", len(s.AllClients))
	}
}
