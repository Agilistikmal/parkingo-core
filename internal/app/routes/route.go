package routes

import (
	"github.com/agilistikmal/parkingo-core/internal/app/controllers"
	"github.com/agilistikmal/parkingo-core/internal/app/middlewares"
	"github.com/gofiber/fiber/v2"
)

type Route struct {
	FiberApp          *fiber.App
	AuthMiddleware    *middlewares.AuthMiddleware
	AuthController    *controllers.AuthController
	UserController    *controllers.UserController
	ParkingController *controllers.ParkingController
}

func NewRoute(fiberApp *fiber.App, authMiddleware *middlewares.AuthMiddleware, authController *controllers.AuthController, userController *controllers.UserController, parkingController *controllers.ParkingController) *Route {
	return &Route{
		FiberApp:          fiberApp,
		AuthMiddleware:    authMiddleware,
		AuthController:    authController,
		UserController:    userController,
		ParkingController: parkingController,
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
	parkingRoutes.Get("/:id", r.ParkingController.GetParking)
	parkingRoutes.Post("/", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.ParkingController.CreateParking)
	parkingRoutes.Patch("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.ParkingController.UpdateParking)
	parkingRoutes.Delete("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.ParkingController.DeleteParking)
}
