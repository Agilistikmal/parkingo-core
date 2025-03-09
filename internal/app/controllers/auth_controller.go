package controllers

import (
	"strings"
	"time"

	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/pkg"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	JWTService  *services.JWTService
	AuthService *services.AuthService
	UserService *services.UserService
}

func NewAuthController(jwtService *services.JWTService, authService *services.AuthService, userService *services.UserService) *AuthController {
	return &AuthController{
		JWTService:  jwtService,
		AuthService: authService,
		UserService: userService,
	}
}

func (c *AuthController) Authenticate(ctx *fiber.Ctx) error {
	googleAuthURL := c.AuthService.GetGoogleAuthURL()
	return ctx.Redirect(googleAuthURL)
}

func (c *AuthController) AuthenticateCallback(ctx *fiber.Ctx) error {
	code := ctx.Query("code")
	if code == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing code",
		})
	}

	token, err := c.AuthService.GetGoogleToken(code)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get token",
		})
	}

	userInfo, err := c.AuthService.GetGoogleUserInfo(token)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get user info",
		})
	}

	logrus.Info(userInfo)

	user, err := c.UserService.GetUserByEmail(userInfo["email"].(string))
	if err != nil {
		username := strings.Split(userInfo["email"].(string), "@")[0]
		user, err = c.UserService.GetUserByUsername(username)
		if user != nil {
			username += "-" + pkg.RandomString(4)
		}
		user, err = c.UserService.CreateUser(&models.CreateUserRequest{
			Email:    userInfo["email"].(string),
			FullName: userInfo["name"].(string),
			Username: username,
			GoogleID: userInfo["id"].(string),
		})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create user",
			})
		}
	}

	tokenString, err := c.JWTService.GenerateToken(user.ID, time.Hour*24)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate token",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"token": tokenString,
		},
	})
}
