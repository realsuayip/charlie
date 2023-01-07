package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"strings"
)

type Handler struct {
	mongoClient *mongo.Client
	database    *mongo.Database
	validate    *validator.Validate
}

func NewHandler() *Handler {
	/*
		A handler is used as receiver in controllers so
		that commonly used instances such as the database
		connector and 'validator' are readily available.
	*/
	mi := MongoConnect()

	validate := validator.New()
	// Register a tag name function so that the fields json
	// annotation can be used in the error messages.
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Handler{
		mongoClient: mi.Client,
		database:    mi.Database,
		validate:    validate,
	}
}

func (h *Handler) validateStruct(s interface{}) fiber.Map {
	// Validates a struct, if any errors returns a map
	// detailing the error with relevant tags.
	err := h.validate.Struct(s)

	if err == nil {
		return nil
	}

	errors := make(fiber.Map)
	for _, err := range err.(validator.ValidationErrors) {
		errors[err.Field()] = fiber.Map{
			"tag": err.Tag(),
		}
	}

	return fiber.Map{
		"detail": "One or more of the fields are invalid.",
		"fields": errors,
	}
}
