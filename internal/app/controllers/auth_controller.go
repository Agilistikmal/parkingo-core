package controllers

import (
	"fmt"
	"net/url"
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
	redirectUrl := ctx.Query("redirect_url")
	if redirectUrl == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing redirect_url",
		})
	}

	state := url.QueryEscape(redirectUrl)
	googleAuthURL := c.AuthService.GetGoogleAuthURL(state)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"url": googleAuthURL,
		},
	})
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

	user, err := c.UserService.GetUserByEmail(userInfo["email"].(string))
	if err != nil {
		username := strings.Split(userInfo["email"].(string), "@")[0]
		user, err = c.UserService.GetUserByUsername(username)
		if user != nil {
			username += "-" + pkg.RandomString(4)
		}
		user, err = c.UserService.CreateUser(&models.CreateUserRequest{
			Email:     userInfo["email"].(string),
			FullName:  userInfo["name"].(string),
			Username:  username,
			AvatarURL: userInfo["picture"].(string),
			GoogleID:  userInfo["id"].(string),
		})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create user",
			})
		}
	}

	if user.AvatarURL == "" {
		_, err = c.UserService.UpdateUser(user.ID, &models.UpdateUserRequest{
			AvatarURL: userInfo["picture"].(string),
		})
		if err != nil {
			logrus.Warn("Failed to update user avatar URL: ", err)
		}
	}

	tokenString, err := c.JWTService.GenerateToken(user.ID, time.Hour*24)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate token",
		})
	}

	state := ctx.Query("state")
	unescapeState, err := url.QueryUnescape(state)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid redirect_url",
		})
	}

	redirectUrl := fmt.Sprintf("%s?token=%s", unescapeState, tokenString)

	return ctx.Redirect(redirectUrl, fiber.StatusFound)
}

func (c *AuthController) VerifyToken(ctx *fiber.Ctx) error {
	bearerToken := ctx.Get("Authorization")
	if bearerToken == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Missing authorization token",
		})
	}

	if len(bearerToken) < 7 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid authorization token",
		})
	}

	token := bearerToken[7:]

	_, err := c.JWTService.ValidateToken(token)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Token is valid",
	})
}
