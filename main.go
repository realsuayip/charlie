package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
)

func main() {
	mi := MongoConnect()

	app := fiber.New()
	app.Use(logger.New())

	// Routes
	app.Get("/contracts/", mi.GetContracts)
	app.Get("/contracts/:id", mi.GetContract)

	err := app.Listen(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
