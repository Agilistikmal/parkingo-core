package controllers

import (
	"encoding/json"
	"time"

	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/queues"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
)

type WebSocketController struct {
	ScannerMQTT *queues.ScannerMQTT
}

func NewWebSocketController(scannerMQTT *queues.ScannerMQTT) *WebSocketController {
	return &WebSocketController{
		ScannerMQTT: scannerMQTT,
	}
}

// HandleDeviceStream handles WebSocket connections for streaming ESP device images
func (c *WebSocketController) HandleDeviceStream(ctx *fiber.Ctx) error {
	// Check if esp_hmac query param exists
	espHmac := ctx.Query("esp_hmac")
	if espHmac == "" {
		logrus.Warn("WebSocket connection attempt without esp_hmac parameter")
		return fiber.NewError(fiber.StatusBadRequest, "Missing esp_hmac query parameter")
	}

	// Get token from query parameter if available
	token := ctx.Query("token")
	if token != "" {
		// Store token in locals for authentication middleware
		ctx.Locals("query_token", token)
		logrus.Infof("WebSocket connection with token query parameter (length: %d)", len(token))
	}

	// Log headers for debugging
	logrus.Infof("WebSocket connection headers for device %s: %+v", espHmac, ctx.GetReqHeaders())

	// Ensure the connection is upgraded to WebSocket
	if !websocket.IsWebSocketUpgrade(ctx) {
		logrus.Warnf("Non-WebSocket request to device stream endpoint for ESP MAC: %s", espHmac)
		return fiber.NewError(fiber.StatusUpgradeRequired, "WebSocket upgrade required")
	}

	logrus.Infof("Preparing WebSocket upgrade for ESP MAC: %s", espHmac)

	// Store ESP MAC in locals for use in the WebSocket handler
	ctx.Locals("esp_hmac", espHmac)
	ctx.Locals("mode", "specific")

	// Return next to upgrade the connection
	return ctx.Next()
}

// HandleAllDevicesStream handles WebSocket connections for streaming all ESP devices
func (c *WebSocketController) HandleAllDevicesStream(ctx *fiber.Ctx) error {
	// Log headers for debugging
	logrus.Infof("WebSocket connection headers for all devices: %+v", ctx.GetReqHeaders())

	// Get token from query parameter
	token := ctx.Query("token")
	if token != "" {
		// Store token in locals for authentication middleware
		ctx.Locals("query_token", token)
		logrus.Infof("WebSocket connection with token query parameter (length: %d)", len(token))
	}

	// Ensure the connection is upgraded to WebSocket
	if !websocket.IsWebSocketUpgrade(ctx) {
		logrus.Warn("Non-WebSocket request to 'all devices' stream endpoint")
		return fiber.NewError(fiber.StatusUpgradeRequired, "WebSocket upgrade required")
	}

	logrus.Info("Preparing WebSocket upgrade for ALL devices")

	// Mark as "all devices" mode
	ctx.Locals("mode", "all")

	// Return next to upgrade the connection
	return ctx.Next()
}

// HandleWebSocketConnection handles the actual WebSocket connection after upgrade
func (c *WebSocketController) HandleWebSocketConnection(conn *websocket.Conn) {
	// Get connection mode
	mode, ok := conn.Locals("mode").(string)
	if !ok {
		logrus.Error("Failed to get connection mode from context")
		conn.Close()
		return
	}

	logrus.Infof("WebSocket connection established with mode: %s", mode)

	// Create a channel for sending messages to this client
	sendChan := make(chan []byte, 20) // Increased buffer size

	// Track connection time for logging
	connectedAt := time.Now()

	// Create a channel to signal when the connection is closed
	done := make(chan struct{})

	var espHmac string
	if mode == "specific" {
		// Handle specific device connection
		espHmac, ok = conn.Locals("esp_hmac").(string)
		if !ok {
			logrus.Error("Failed to get ESP MAC from context")
			conn.Close()
			return
		}

		logrus.Infof("Registering WebSocket client for ESP MAC: %s", espHmac)

		// Register the client with the MQTT handler
		c.ScannerMQTT.RegisterSpecificClient(espHmac, sendChan)

		// Send an initial ping message to confirm connection is working
		pingMsg := map[string]interface{}{
			"type":      "connection_status",
			"connected": true,
			"esp_hmac":  espHmac,
			"timestamp": time.Now().UnixMilli(),
			"message":   "WebSocket connection established",
		}
		pingData, _ := json.Marshal(pingMsg)

		if err := conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
			logrus.Errorf("Error sending initial ping: %v", err)
		} else {
			logrus.Info("Sent initial ping message to confirm WebSocket connection is working")
		}

		// Get current image and send it immediately if available
		deviceImage := c.ScannerMQTT.GetDeviceImage(espHmac)
		if deviceImage != nil {
			if data, err := json.Marshal(deviceImage); err == nil {
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logrus.Errorf("Error sending initial image data: %v", err)
				} else {
					logrus.Infof("Sent initial image data for ESP MAC: %s directly", espHmac)
				}
			}
		}

		// Send a test broadcast after 5 seconds to verify channel is working
		go func() {
			time.Sleep(5 * time.Second)
			c.ScannerMQTT.BroadcastTestMessage(espHmac, true)
			logrus.Infof("Triggered test broadcast for specific device: %s", espHmac)
		}()

		// Clean up when the connection is closed
		defer func() {
			logrus.Infof("Unregistering WebSocket client for ESP MAC: %s (was connected for %v)",
				espHmac, time.Since(connectedAt))
			c.ScannerMQTT.UnregisterClient(espHmac, sendChan)
			close(sendChan)
			close(done)
			conn.Close()
		}()
	} else if mode == "all" {
		// Handle "all devices" connection
		logrus.Info("Registering WebSocket client for ALL devices")

		// Register the client with the MQTT handler
		c.ScannerMQTT.RegisterAllClient(sendChan)

		// Send an initial ping message to confirm connection is working
		pingMsg := map[string]interface{}{
			"type":      "connection_status",
			"connected": true,
			"mode":      "all_devices",
			"timestamp": time.Now().UnixMilli(),
			"message":   "WebSocket connection established for all devices",
		}
		pingData, _ := json.Marshal(pingMsg)

		if err := conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
			logrus.Errorf("Error sending initial ping: %v", err)
		} else {
			logrus.Info("Sent initial ping message to confirm WebSocket connection is working")
		}

		// Get all current devices and send immediately
		allDevices := c.ScannerMQTT.GetAllDevices()
		if len(allDevices) > 0 {
			if data, err := json.Marshal(allDevices); err == nil {
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logrus.Errorf("Error sending initial devices data: %v", err)
				} else {
					logrus.Infof("Sent initial data for %d devices directly", len(allDevices))
				}
			}
		}

		// Send a test broadcast after 5 seconds to verify channel is working
		go func() {
			time.Sleep(5 * time.Second)
			c.ScannerMQTT.BroadcastTestMessage("TEST-ESP-ALL", true)
			logrus.Info("Triggered test broadcast for all devices")
		}()

		// Clean up when the connection is closed
		defer func() {
			logrus.Infof("Unregistering WebSocket client for ALL devices (was connected for %v)",
				time.Since(connectedAt))
			c.ScannerMQTT.UnregisterClient("", sendChan) // Empty ESP MAC means "all devices"
			close(sendChan)
			close(done)
			conn.Close()
		}()
	} else {
		logrus.Errorf("Unknown WebSocket connection mode: %s", mode)
		conn.Close()
		return
	}

	// Set up a ticker for periodic pings to keep connection alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Set up a refresh ticker to periodically resend current data
	// This ensures client gets updates even if broadcast mechanism fails
	refreshTicker := time.NewTicker(15 * time.Second)
	defer refreshTicker.Stop()

	// Start a goroutine to listen for messages to send to the client
	msgCounter := 0
	go func() {
		for {
			select {
			case data, ok := <-sendChan:
				if !ok {
					logrus.Info("Send channel closed, stopping WebSocket sender goroutine")
					return
				}

				msgCounter++
				logrus.Infof("Received data to send over WebSocket (message #%d, %d bytes)",
					msgCounter, len(data))

				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logrus.Errorf("Error writing to WebSocket: %v", err)
					// Signal that we should stop
					select {
					case <-done:
						// Already closed
					default:
						close(done)
					}
					return
				}

				logrus.Infof("Successfully sent message #%d over WebSocket", msgCounter)
			case <-done:
				return
			case <-pingTicker.C:
				// Send a ping to keep the connection alive
				keepAliveMsg := map[string]interface{}{
					"type":      "keep_alive",
					"timestamp": time.Now().UnixMilli(),
				}

				if mode == "specific" && espHmac != "" {
					keepAliveMsg["esp_hmac"] = espHmac
				} else {
					keepAliveMsg["mode"] = "all_devices"
				}

				pingData, _ := json.Marshal(keepAliveMsg)

				if err := conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
					logrus.Errorf("Error sending keep-alive ping: %v", err)
					return
				}
				logrus.Debug("Sent keep-alive ping message")
			case <-refreshTicker.C:
				// Periodically resend current data to ensure client has latest images
				// but don't send if the last message was very recent
				now := time.Now().UnixMilli()

				if mode == "specific" && espHmac != "" {
					// Get latest image for this ESP
					deviceImage := c.ScannerMQTT.GetDeviceImage(espHmac)
					if deviceImage != nil {
						// Create a copy with an updated timestamp to show it's a refresh
						refreshedImage := &models.ParkingImage{
							ESPHmac:   deviceImage.ESPHmac,
							ImageData: deviceImage.ImageData,
							Timestamp: now,
						}

						if data, err := json.Marshal(refreshedImage); err == nil {
							// Only update timestamp in the message, not in the stored data
							if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
								logrus.Errorf("Error sending refresh image data: %v", err)
							} else {
								logrus.Infof("Sent refresh image data with updated timestamp for ESP MAC: %s", espHmac)
							}
						}
					}
				} else if mode == "all" {
					// Get all latest devices
					allDevices := c.ScannerMQTT.GetAllDevices()
					if len(allDevices) > 0 {
						// Create copies with updated timestamps
						refreshedDevices := make([]*models.ParkingImage, len(allDevices))
						for i, device := range allDevices {
							refreshedDevices[i] = &models.ParkingImage{
								ESPHmac:   device.ESPHmac,
								ImageData: device.ImageData,
								Timestamp: now,
							}
						}

						if data, err := json.Marshal(refreshedDevices); err == nil {
							if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
								logrus.Errorf("Error sending refresh devices data: %v", err)
							} else {
								logrus.Infof("Sent refresh data with updated timestamps for %d devices", len(refreshedDevices))
							}
						}
					}
				}
			}
		}
	}()

	// Listen for WebSocket messages from the client
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			logrus.Infof("WebSocket connection closed: %v", err)
			// Signal the sender goroutine to stop
			select {
			case <-done:
				// Already closed
			default:
				close(done)
			}
			break
		}

		logrus.Infof("Received WebSocket message from client, type: %d, length: %d bytes",
			messageType, len(msg))

		// Handle client ping (if needed)
		if messageType == websocket.PingMessage {
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				logrus.Error("Error sending pong: ", err)
				break
			}
			logrus.Info("Sent pong response to client ping")
		}
	}
}

// GetAllDevices returns a list of all ESP devices that have sent data
func (c *WebSocketController) GetAllDevices(ctx *fiber.Ctx) error {
	logrus.Info("REST API request: GetAllDevices")
	devices := c.ScannerMQTT.GetAllDevices()
	logrus.Infof("Returning %d devices via REST API", len(devices))
	return ctx.JSON(fiber.Map{
		"status":  "success",
		"count":   len(devices),
		"devices": devices,
	})
}

// GetDeviceImage returns the latest image for a specific ESP device
func (c *WebSocketController) GetDeviceImage(ctx *fiber.Ctx) error {
	espHmac := ctx.Params("esp_hmac")
	if espHmac == "" {
		logrus.Warn("REST API request: GetDeviceImage without esp_hmac parameter")
		return fiber.NewError(fiber.StatusBadRequest, "Missing ESP MAC parameter")
	}

	logrus.Infof("REST API request: GetDeviceImage for ESP MAC: %s", espHmac)
	deviceImage := c.ScannerMQTT.GetDeviceImage(espHmac)
	if deviceImage == nil {
		logrus.Warnf("REST API: No image data found for ESP MAC: %s", espHmac)
		return fiber.NewError(fiber.StatusNotFound, "No image data available for this device")
	}

	logrus.Infof("REST API: Returning image for ESP MAC: %s", espHmac)
	return ctx.JSON(deviceImage)
}
