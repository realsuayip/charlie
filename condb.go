package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

var DBName = os.Getenv("DATABASE_NAME")
var MongoURI = os.Getenv("MONGO_URI")

type MongoInstance struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func MongoConnect() *MongoInstance {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(MongoURI).SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(DBName)
	return &MongoInstance{Client: client, Database: db}
}
