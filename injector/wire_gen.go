// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package injector

import (
	"github.com/agilistikmal/parkingo-core/internal/app/controllers"
	"github.com/agilistikmal/parkingo-core/internal/app/middlewares"
	"github.com/agilistikmal/parkingo-core/internal/app/routes"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/database"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/fiberapp"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/validation"
)

// Injectors from injector.go:

func InjectRoutes() *routes.Route {
	app := fiberapp.NewFiberApp()
	jwtService := services.NewJWTService()
	db := database.NewDatabase()
	validate := validation.New()
	userService := services.NewUserService(db, validate)
	authMiddleware := middlewares.NewAuthMiddleware(jwtService, userService)
	userController := controllers.NewUserController(userService)
	route := routes.NewRoute(app, authMiddleware, userController)
	return route
}
