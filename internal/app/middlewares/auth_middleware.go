package middlewares

import (
	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/gofiber/fiber/v2"
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
