package controllers

import (
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
		return fiber.NewError(fiber.StatusBadRequest, "Missing esp_hmac query parameter")
	}

	// Ensure the connection is upgraded to WebSocket
	if !websocket.IsWebSocketUpgrade(ctx) {
		return fiber.NewError(fiber.StatusUpgradeRequired, "WebSocket upgrade required")
	}

	// Store ESP MAC in locals for use in the WebSocket handler
	ctx.Locals("esp_hmac", espHmac)
	ctx.Locals("mode", "specific")

	// Return next to upgrade the connection
	return ctx.Next()
}

// HandleAllDevicesStream handles WebSocket connections for streaming all ESP devices
func (c *WebSocketController) HandleAllDevicesStream(ctx *fiber.Ctx) error {
	// Ensure the connection is upgraded to WebSocket
	if !websocket.IsWebSocketUpgrade(ctx) {
		return fiber.NewError(fiber.StatusUpgradeRequired, "WebSocket upgrade required")
	}

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

	// Create a channel for sending messages to this client
	sendChan := make(chan []byte, 10)

	if mode == "specific" {
		// Handle specific device connection
		espHmac, ok := conn.Locals("esp_hmac").(string)
		if !ok {
			logrus.Error("Failed to get ESP MAC from context")
			conn.Close()
			return
		}

		// Register the client with the MQTT handler
		c.ScannerMQTT.RegisterSpecificClient(espHmac, sendChan)

		// Clean up when the connection is closed
		defer func() {
			c.ScannerMQTT.UnregisterClient(espHmac, sendChan)
			close(sendChan)
			conn.Close()
		}()
	} else if mode == "all" {
		// Handle "all devices" connection
		c.ScannerMQTT.RegisterAllClient(sendChan)

		// Clean up when the connection is closed
		defer func() {
			c.ScannerMQTT.UnregisterClient("", sendChan) // Empty ESP MAC means "all devices"
			close(sendChan)
			conn.Close()
		}()
	} else {
		logrus.Errorf("Unknown WebSocket connection mode: %s", mode)
		conn.Close()
		return
	}

	// Start a goroutine to listen for messages to send to the client
	go func() {
		for data := range sendChan {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				logrus.Errorf("Error writing to WebSocket: %v", err)
				return
			}
		}
	}()

	// Listen for WebSocket messages from the client
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			logrus.Infof("WebSocket connection closed: %v", err)
			break
		}

		// Handle client ping (if needed)
		if messageType == websocket.PingMessage {
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				logrus.Error("Error sending pong: ", err)
				break
			}
		}
	}
}

// GetAllDevices returns a list of all ESP devices that have sent data
func (c *WebSocketController) GetAllDevices(ctx *fiber.Ctx) error {
	devices := c.ScannerMQTT.GetAllDevices()
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
		return fiber.NewError(fiber.StatusBadRequest, "Missing ESP MAC parameter")
	}

	deviceImage := c.ScannerMQTT.GetDeviceImage(espHmac)
	if deviceImage == nil {
		return fiber.NewError(fiber.StatusNotFound, "No image data available for this device")
	}

	return ctx.JSON(deviceImage)
}
