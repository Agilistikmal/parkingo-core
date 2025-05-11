package middlewares

import (
	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	JWTService  *services.JWTService
	UserService *services.UserService
}

func NewAuthMiddleware(jwtService *services.JWTService, userService *services.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		JWTService:  jwtService,
		UserService: userService,
	}
}

func (m *AuthMiddleware) VerifyAuthencitated(c *fiber.Ctx) error {
	bearerToken := c.Get("Authorization")
	if bearerToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Missing authorization token",
		})
	}

	if len(bearerToken) < 7 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid authorization token",
		})
	}

	token := bearerToken[7:]

	userID, err := m.JWTService.GetUserIDFromToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid authorization token",
		})
	}

	user, err := m.UserService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	c.Locals("user", user)

	return c.Next()
}

func (m *AuthMiddleware) VerifyAdminAccess(c *fiber.Ctx) error {
	authUser := c.Locals("user").(*models.User)
	if !authUser.IsAdmin() {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized to access this resource",
		})
	}

	return c.Next()
}

// VerifyWebSocketAuth checks for authentication via Authorization header OR token query parameter
// This is specifically for WebSocket routes that need to support both authentication methods
func (m *AuthMiddleware) VerifyWebSocketAuth(c *fiber.Ctx) error {
	var token string
	var isQueryToken bool

	// First check if we have a token in query parameter (stored from WebSocket handler)
	queryToken, ok := c.Locals("query_token").(string)
	if ok && queryToken != "" {
		token = queryToken
		isQueryToken = true
		logrus.Infof("Using token from query parameter for WebSocket authentication")
	} else {
		// Fall back to checking Authorization header
		bearerToken := c.Get("Authorization")
		if bearerToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Missing authorization token",
			})
		}

		if len(bearerToken) < 7 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid authorization token",
			})
		}

		token = bearerToken[7:]
	}

	// Validate token and get user
	userID, err := m.JWTService.GetUserIDFromToken(token)
	if err != nil {
		logrus.Warnf("Invalid token for WebSocket authentication: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid authorization token",
		})
	}

	user, err := m.UserService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	// Set user in context locals
	c.Locals("user", user)
	if isQueryToken {
		logrus.Infof("WebSocket authenticated via query token for user: %s (ID: %d)", user.Username, user.ID)
	}

	return c.Next()
}
