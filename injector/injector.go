//go:build wireinject
// +build wireinject

package injector

import (
	"github.com/agilistikmal/parkingo-core/internal/app/controllers"
	"github.com/agilistikmal/parkingo-core/internal/app/middlewares"
	"github.com/agilistikmal/parkingo-core/internal/app/routes"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/database"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/fiberapp"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/validation"
	"github.com/google/wire"
)

func InjectRoutes() *routes.Route {
	wire.Build(
		fiberapp.NewFiberApp,
		database.NewDatabase,
		validation.New,
		services.NewAuthService,
		services.NewUserService,
		services.NewJWTService,
		controllers.NewAuthController,
		controllers.NewUserController,
		middlewares.NewAuthMiddleware,
		routes.NewRoute,
	)

	return nil
}
