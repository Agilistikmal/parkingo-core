package routes

import (
	"encoding/json"
	"time"

	"github.com/agilistikmal/parkingo-core/internal/app/controllers"
	"github.com/agilistikmal/parkingo-core/internal/app/jobs"
	"github.com/agilistikmal/parkingo-core/internal/app/middlewares"
	"github.com/agilistikmal/parkingo-core/internal/app/queues"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
)

type Route struct {
	FiberApp            *fiber.App
	AuthMiddleware      *middlewares.AuthMiddleware
	AuthController      *controllers.AuthController
	UserController      *controllers.UserController
	ParkingController   *controllers.ParkingController
	BookingController   *controllers.BookingController
	WebSocketController *controllers.WebSocketController
	BookingJob          *jobs.BookingJob
	ScannerMQTT         *queues.ScannerMQTT
}

func NewRoute(
	fiberApp *fiber.App,
	authMiddleware *middlewares.AuthMiddleware,
	authController *controllers.AuthController,
	userController *controllers.UserController,
	parkingController *controllers.ParkingController,
	bookingController *controllers.BookingController,
	webSocketController *controllers.WebSocketController,
	bookingJob *jobs.BookingJob,
	scannerMQTT *queues.ScannerMQTT,
) *Route {
	return &Route{
		FiberApp:            fiberApp,
		AuthMiddleware:      authMiddleware,
		AuthController:      authController,
		UserController:      userController,
		ParkingController:   parkingController,
		BookingController:   bookingController,
		WebSocketController: webSocketController,
		BookingJob:          bookingJob,
		ScannerMQTT:         scannerMQTT,
	}
}

func (r *Route) RegisterRoutes() {
	v1 := r.FiberApp.Group("/v1")

	authRoutes := v1.Group("/authenticate")
	authRoutes.Get("/", r.AuthController.Authenticate)
	authRoutes.Get("/callback", r.AuthController.AuthenticateCallback)

	userRoutes := v1.Group("/users")
	// USER ROLE
	userRoutes.Get("/me", r.AuthMiddleware.VerifyAuthencitated, r.UserController.GetCurrentUser)
	userRoutes.Patch("/me", r.AuthMiddleware.VerifyAuthencitated, r.UserController.UpdateCurrentUser)
	userRoutes.Delete("/me", r.AuthMiddleware.VerifyAuthencitated, r.UserController.DeleteCurrentUser)
	// ADMIN ROLE
	userRoutes.Get("/", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.GetAllUsers)
	userRoutes.Get("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.GetUserByID)
	userRoutes.Post("/", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.CreateUser)
	userRoutes.Patch("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.UpdateUser)
	userRoutes.Delete("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.DeleteUser)

	parkingRoutes := v1.Group("/parkings")
	parkingRoutes.Get("/", r.ParkingController.GetParkings)
	parkingRoutes.Get("/:id", r.ParkingController.GetParkingByID)
	parkingRoutes.Get("/slug/:slug", r.ParkingController.GetParkingBySlug)
	parkingRoutes.Post("/", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.ParkingController.CreateParking)
	parkingRoutes.Patch("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.ParkingController.UpdateParking)
	parkingRoutes.Delete("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.ParkingController.DeleteParking)

	bookingRoutes := v1.Group("/bookings")
	bookingRoutes.Get("/", r.AuthMiddleware.VerifyAuthencitated, r.BookingController.GetBookings)
	bookingRoutes.Get("/:id", r.AuthMiddleware.VerifyAuthencitated, r.BookingController.GetBookingByID)
	bookingRoutes.Get("/reference/:reference", r.AuthMiddleware.VerifyAuthencitated, r.BookingController.GetBookingByReference)
	bookingRoutes.Post("/", r.AuthMiddleware.VerifyAuthencitated, r.BookingController.CreateBooking)
	bookingRoutes.Post("/callback/payment", r.BookingController.PaymentCallback)
	bookingRoutes.Patch("/:id", r.AuthMiddleware.VerifyAuthencitated, r.BookingController.UpdateBooking)
	bookingRoutes.Delete("/:id", r.AuthMiddleware.VerifyAuthencitated, r.BookingController.DeleteBooking)
	bookingRoutes.Post("/validate", r.BookingController.ValidateBooking)

	// ESP devices routes - for admin to monitor devices
	deviceRoutes := v1.Group("/devices")
	deviceRoutes.Get("/", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.WebSocketController.GetAllDevices)
	deviceRoutes.Get("/:esp_hmac", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.WebSocketController.GetDeviceImage)

	// WebSocket middleware
	r.FiberApp.Use("/ws", func(c *fiber.Ctx) error {
		logrus.Infof("WebSocket middleware handling request: %s %s", c.Method(), c.Path())
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol
		if websocket.IsWebSocketUpgrade(c) {
			logrus.Info("WebSocket upgrade requested, proceeding")
			return c.Next()
		}
		logrus.Warn("Non-WebSocket request to WebSocket endpoint, returning Upgrade Required")
		return fiber.ErrUpgradeRequired
	})

	// WebSocket test endpoint that sends a message every second
	r.FiberApp.Get("/ws/test", func(c *fiber.Ctx) error {
		logrus.Info("WebSocket test endpoint called")
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}, websocket.New(func(c *websocket.Conn) {
		logrus.Info("WebSocket TEST connection established")

		// Send a welcome message
		welcomeMsg := map[string]interface{}{
			"type":      "welcome",
			"message":   "WebSocket test connection successful!",
			"timestamp": time.Now().UnixMilli(),
		}
		welcomeData, _ := json.Marshal(welcomeMsg)
		if err := c.WriteMessage(websocket.TextMessage, welcomeData); err != nil {
			logrus.Errorf("Failed to send welcome message: %v", err)
			return
		}

		// Send periodic messages until client disconnects
		count := 0
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		defer func() {
			logrus.Info("WebSocket TEST connection closed")
		}()

		// Create a channel to signal when the connection is closed
		done := make(chan struct{})

		// Start a goroutine to read messages
		go func() {
			defer close(done)
			for {
				_, _, err := c.ReadMessage()
				if err != nil {
					logrus.Infof("WebSocket TEST client disconnected: %v", err)
					return
				}
			}
		}()

		for {
			select {
			case <-ticker.C:
				count++
				pingMsg := map[string]interface{}{
					"type":      "ping",
					"count":     count,
					"timestamp": time.Now().UnixMilli(),
					"message":   "This is a test message from the server",
				}
				pingData, _ := json.Marshal(pingMsg)

				if err := c.WriteMessage(websocket.TextMessage, pingData); err != nil {
					logrus.Errorf("Error writing to WebSocket: %v", err)
					return
				}
				logrus.Infof("Sent test message #%d to client", count)
			case <-done:
				return
			}
		}
	}))

	// ESP MAC-based device stream WebSocket
	r.FiberApp.Get("/ws/device", r.WebSocketController.HandleDeviceStream, r.AuthMiddleware.VerifyWebSocketAuth, r.AuthMiddleware.VerifyAdminAccess, websocket.New(r.WebSocketController.HandleWebSocketConnection))

	// All devices stream WebSocket (admin only) - with token query param support
	r.FiberApp.Get("/ws/devices/all", r.WebSocketController.HandleAllDevicesStream, r.AuthMiddleware.VerifyWebSocketAuth, r.AuthMiddleware.VerifyAdminAccess, websocket.New(r.WebSocketController.HandleWebSocketConnection))
}
