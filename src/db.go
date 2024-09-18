package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

func mongoDbQueryTest() {
	var result bson.M
	err := db.Collection("movies").
		FindOne(context.TODO(), bson.D{{"title", "Back to the Future"}}).
		Decode(&result)

	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with title")
		return
	}
	if err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", jsonData)
}

func initMongoDb() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. ")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	defer func() {
		err := client.Disconnect(context.TODO())
		if err != nil {
			panic(err)
		}
	}()

	db = client.Database("sample_mflix")
}
