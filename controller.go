package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (h *Handler) CreateContract(c *fiber.Ctx) error {
	payload := new(struct {
		StartAt time.Time `json:"start_at" validate:"required"`
		EndAt   time.Time `json:"end_at" validate:"required"`
		Meta    fiber.Map `json:"meta" validate:"required"`
		Data    fiber.Map `json:"data" validate:"required"`
	})

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"detail": err.Error()})
	}

	if fieldErrors := h.validateStruct(payload); fieldErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fieldErrors)
	}

	contract, err := NewContract(payload.StartAt, payload.EndAt, payload.Data, payload.Meta)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"detail": err.Error()})
	}
	contract.UpdatedAt = time.Now().UTC()

	coll := h.database.Collection("contract")
	if _, err := coll.InsertOne(context.TODO(), contract); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"detail": err.Error()})
	}
	return c.Status(201).JSON(contract)
}

func getContract(mc *mongo.Collection, c *fiber.Ctx) (*Contract, error) {
	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var contract *Contract
	filter := bson.M{"_id": objectID}
	err = mc.FindOne(context.TODO(), filter).Decode(&contract)

	if err != nil {
		return nil, err
	}
	return contract, nil
}

func (h *Handler) GetContract(c *fiber.Ctx) error {
	coll := h.database.Collection("contract")
	contract, err := getContract(coll, c)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return c.JSON(contract)
}

func (h *Handler) ListContracts(c *fiber.Ctx) error {
	var filter bson.M
	if cursor := c.Query("cursor"); cursor != "" {
		objectID, err := primitive.ObjectIDFromHex(cursor)
		if err == nil {
			// $lt since the query is ordered by {_id, -1}
			filter = bson.M{"_id": bson.M{"$lt": objectID}}
		}
	}

	opts := options.Find().
		SetProjection(bson.D{{Key: "items", Value: 0}}).
		SetLimit(10).
		SetSort(bson.D{{Key: "_id", Value: -1}})

	collection := h.database.Collection("contract")
	cur, err := collection.Find(context.TODO(), filter, opts)

	contracts := make([]fiber.Map, 0)
	if err = cur.All(context.TODO(), &contracts); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	results := fiber.Map{"results": contracts}
	return c.JSON(results)
}

func (h *Handler) BranchContract(c *fiber.Ctx) error {
	payload := new(struct {
		StartAt time.Time `json:"start_at" validate:"required"`
		EndAt   time.Time `json:"end_at"`
		Data    fiber.Map `json:"data" validate:"required"`
	})

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"detail": err.Error()})
	}

	if fieldErrors := h.validateStruct(payload); fieldErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fieldErrors)
	}

	coll := h.database.Collection("contract")
	contract, err := getContract(coll, c)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	_, err = contract.Branch(payload.StartAt, payload.EndAt, payload.Data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"detail": err.Error()})
	}

	filter := bson.M{"_id": contract.ID}
	update := bson.M{"$set": bson.M{"items": contract.Items}, "$currentDate": bson.M{"updated_at": true}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var document *Contract
	if err = coll.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&document); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"detail": err.Error()})
	}
	return c.JSON(document)
}
