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
	mi := MongoConnect()

	validate := validator.New()
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
