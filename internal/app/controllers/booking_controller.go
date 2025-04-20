package controllers

import (
	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/pkg"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/gofiber/fiber/v2"
)

type BookingController struct {
	BookingService *services.BookingService
	ParkingService *services.ParkingService
	UserService    *services.UserService
}

func NewBookingController(bookingService *services.BookingService, parkingService *services.ParkingService, userService *services.UserService) *BookingController {
	return &BookingController{
		BookingService: bookingService,
		ParkingService: parkingService,
		UserService:    userService,
	}
}

func (c *BookingController) GetBookings(ctx *fiber.Ctx) error {
	bookings, err := c.BookingService.GetBookings()
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": bookings,
	})
}

func (c *BookingController) GetBookingByID(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid booking ID",
		})
	}

	booking, err := c.BookingService.GetBookingByID(id)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": booking,
	})
}

func (c *BookingController) GetBookingByReference(ctx *fiber.Ctx) error {
	reference := ctx.Params("reference")
	if reference == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid booking reference",
		})
	}
	booking, err := c.BookingService.GetBookingByReference(reference)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": booking,
	})
}

func (c *BookingController) CreateBooking(ctx *fiber.Ctx) error {
	authUser := ctx.Locals("user").(*models.User)

	var req *models.CreateBookingRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	booking, err := c.BookingService.CreateBooking(authUser.ID, req)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": booking,
	})
}

func (c *BookingController) UpdateBooking(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid booking ID",
		})
	}

	var req *models.UpdateBookingRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	booking, err := c.BookingService.UpdateBooking(id, req)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": booking,
	})
}

func (c *BookingController) DeleteBooking(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid booking ID",
		})
	}

	err = c.BookingService.DeleteBooking(id)
	if err != nil {
		return pkg.HandlerError(ctx, err)
	}

	return ctx.Status(fiber.StatusNoContent).JSON(nil)
}
