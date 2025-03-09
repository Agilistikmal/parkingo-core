package main

import (
	"github.com/agilistikmal/parkingo-core/injector"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/config"
)

func main() {
	config.Load()

	routes := injector.InjectRoutes()
	routes.RegisterRoutes()

	routes.FiberApp.Listen(":3000")
}
