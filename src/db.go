package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func GetUser(filter bson.D) (User, error) {
	var result bson.M
	err := mongoClient.
		Database("GoCharge").
		Collection("Users").
		FindOne(context.TODO(), filter).
		Decode(&result)
	if err != nil {
		return User{}, err
	}

	return FromMongoDoc[User](result)
}

func CreateUser(username string, password string, email string, role string) (string, error) {
	new_user := NewUser{
		Username:           username,
		Password:           password,
		Email:              email,
		Role:               role,
		PhotoURL:           "",
		FavoriteStationIDs: []string{},
	}

	new_user_doc, err := ToMongoDoc(new_user)
	if err != nil {
		return "", err
	}

	result, err := mongoClient.
		Database("GoCharge").
		Collection("Users").
		InsertOne(context.TODO(), new_user_doc)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(result.InsertedID), nil
}

func InitMongoDb() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("No .env file found")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. ")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	userIndexes := client.Database("GoCharge").
		Collection("Users").
		Indexes()

	userIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{"username", 1}},
		Options: options.Index().SetUnique(true),
	})

	mongoClient = client
}
