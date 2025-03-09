package pkg

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func HandlerError(ctx *fiber.Ctx, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Resource not found",
		})
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Duplicated key",
		})
	}

	switch e := err.(type) {
	case validator.ValidationErrors:
		// Handle go-validator error
		// Return a response with the validation errors
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": e.Error(),
		})
	default:
		// Handle other types of errors
		// Return a generic error response
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
	}
}
