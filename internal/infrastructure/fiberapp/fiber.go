package fiberapp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewFiberApp() *fiber.App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://parkingo.agil.zip,http://localhost:3000",
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
	}))

	return app
}
