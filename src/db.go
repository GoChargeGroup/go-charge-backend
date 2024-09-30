package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

// USER WRAPPER FUNCTIONS

func GetUser(filter bson.D) (User, error) {
	return GetOne[User]("Users", filter)
}

func CreateUser(username string, password string, email string, role string) (primitive.ObjectID, error) {
	new_user := NewUser{
		Username:           username,
		Password:           password,
		Email:              email,
		Role:               role,
		PhotoURL:           "",
		FavoriteStationIDs: []string{},
	}
	return CreateOne("Users", new_user)
}

func UpdateUser(filter bson.D, update bson.D) error {
	return UpdateOne("Users", filter, update)
}

func DeleteUser(filter bson.D) error {
	return DeleteOne("Users", filter)
}

// DB WRAPPER FUNCTIONS

func GetOne[T interface{}](collection string, filter bson.D) (T, error) {
	var result T
	err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		FindOne(context.TODO(), filter).
		Decode(&result)
	return result, err
}

func CreateOne(collection string, document interface{}) (primitive.ObjectID, error) {
	result, err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		InsertOne(context.TODO(), document)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

func UpdateOne(collection string, filter bson.D, update bson.D) error {
	_, err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		UpdateOne(context.TODO(), filter, update)
	return err
}

func DeleteOne(collection string, filter bson.D) error {
	_, err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		DeleteOne(context.TODO(), filter)
	return err
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
