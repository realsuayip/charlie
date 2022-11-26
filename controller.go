package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mi *MongoInstance) GetContract(c *fiber.Ctx) error {
	collection := mi.Database.Collection("contract")
	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	var contract *Contract
	filter := bson.M{"_id": objectID}
	err = collection.FindOne(context.TODO(), filter).Decode(&contract)

	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return c.JSON(contract)
}

func (mi *MongoInstance) GetContracts(c *fiber.Ctx) error {
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

	collection := mi.Database.Collection("contract")
	cur, err := collection.Find(context.TODO(), filter, opts)

	contracts := make([]map[string]interface{}, 0)
	if err = cur.All(context.TODO(), &contracts); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	results := map[string]interface{}{
		"results": contracts,
	}
	return c.JSON(results)
}
