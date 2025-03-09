package routes

import (
	"github.com/agilistikmal/parkingo-core/internal/app/controllers"
	"github.com/agilistikmal/parkingo-core/internal/app/middlewares"
	"github.com/gofiber/fiber/v2"
)

type Route struct {
	FiberApp       *fiber.App
	AuthMiddleware *middlewares.AuthMiddleware
	UserController *controllers.UserController
}

func NewRoute(fiberApp *fiber.App, authMiddleware *middlewares.AuthMiddleware, userController *controllers.UserController) *Route {
	return &Route{
		FiberApp:       fiberApp,
		AuthMiddleware: authMiddleware,
		UserController: userController,
	}
}

func (r *Route) RegisterRoutes() {
	v1 := r.FiberApp.Group("/v1")

	userRoutes := v1.Group("/users")
	userRoutes.Get("/me", r.AuthMiddleware.VerifyAuthencitated, r.UserController.GetCurrentUser)
	userRoutes.Get("/", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.GetAllUsers)
	userRoutes.Get("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.GetUserByID)
	userRoutes.Post("/", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.CreateUser)
	userRoutes.Patch("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.UpdateUser)
	userRoutes.Delete("/:id", r.AuthMiddleware.VerifyAuthencitated, r.AuthMiddleware.VerifyAdminAccess, r.UserController.DeleteUser)
}
