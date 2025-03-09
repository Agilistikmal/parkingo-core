package controllers

import (
	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	UserService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

func (c *UserController) GetCurrentUser(ctx *fiber.Ctx) error {
	authUser := ctx.Locals("user").(*models.User)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": authUser,
	})
}

func (c *UserController) GetAllUsers(ctx *fiber.Ctx) error {
	users, err := c.UserService.GetAllUsers()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get users",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": users,
	})
}

func (c *UserController) GetUserByID(ctx *fiber.Ctx) error {
	userID, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	user, err := c.UserService.GetUserByID(userID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": user,
	})
}

func (c *UserController) CreateUser(ctx *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	user, err := c.UserService.CreateUser(&req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create user",
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": user,
	})
}

func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {
	userID, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	var req models.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	user, err := c.UserService.UpdateUser(userID, &req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update user",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": user,
	})
}

func (c *UserController) DeleteUser(ctx *fiber.Ctx) error {
	userID, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	err = c.UserService.DeleteUser(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete user",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(nil)
}
