package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
)

func main() {
	h := NewHandler()

	app := fiber.New()
	app.Use(logger.New())

	// Routes
	app.Post("/contracts/", h.CreateContract)
	app.Get("/contracts/", h.ListContracts)
	app.Get("/contracts/:id", h.GetContract)

	err := app.Listen(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
