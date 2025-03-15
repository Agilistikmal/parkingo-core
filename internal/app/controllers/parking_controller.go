package controllers

import (
	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/pkg"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/gofiber/fiber/v2"
)

type ParkingController struct {
	ParkingService *services.ParkingService
}

func NewParkingController(parkingService *services.ParkingService) *ParkingController {
	return &ParkingController{
		ParkingService: parkingService,
	}
}

func (c *ParkingController) GetParkings(ctx *fiber.Ctx) error {
	parkings, err := c.ParkingService.GetParkings()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get parkings",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": parkings,
	})
}

func (c *ParkingController) GetParkingByID(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid parking ID",
		})
	}

	parking, err := c.ParkingService.GetParkingByID(id)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": parking,
	})
}

func (c *ParkingController) GetParkingBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")

	parking, err := c.ParkingService.GetParkingBySlug(slug)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": parking,
	})
}

func (c *ParkingController) CreateParking(ctx *fiber.Ctx) error {
	authUser := ctx.Locals("user").(*models.User)

	var req *models.CreateParkingRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	parking, err := c.ParkingService.CreateParking(authUser.ID, req)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": parking,
	})
}

func (c *ParkingController) UpdateParking(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid parking ID",
		})
	}

	var req *models.UpdateParkingRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	parking, err := c.ParkingService.UpdateParking(id, req)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": parking,
	})
}

func (c *ParkingController) DeleteParking(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid parking ID",
		})
	}

	err = c.ParkingService.DeleteParking(id)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Parking deleted successfully",
	})
}
