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

// STATION WRAPPER FUNCTIONS

func CreateStation(owner_id primitive.ObjectID, name string, description string, coordinates [2]float64) (primitive.ObjectID, error) {
	new_station := Station{
		ID:          primitive.NewObjectID(),
		OwnerID:     owner_id,
		PictureURLs: []string{},
		Name:        name,
		Description: description,
		Coordinates: coordinates,
		IsPublic:    false,
	}
	return CreateOne("Stations", new_station)
}

func GetStation(filter bson.D) (Station, error) {
	return GetOne[Station]("Stations", filter)
}

func GetStations(filter bson.D, k int64) ([]Station, error) {
	return GetAll[Station]("Stations", filter, k)
}

// STATION WRAPPER FUNCTIONS

func CreateCharger(station_id primitive.ObjectID, name string, description string, kwh_type string, charger_type string, price float64) (primitive.ObjectID, error) {
	new_charger := Charger{
		ID:             primitive.NewObjectID(),
		StationID:      station_id,
		Name:           name,
		Description:    description,
		KWhTypesId:     kwh_type,
		ChargerTypesId: charger_type,
		Price:          price,
		TotalPayments:  0,
	}
	return CreateOne("Chargers", new_charger)
}

func GetCharger(filter bson.D) (Charger, error) {
	return GetOne[Charger]("Chargers", filter)
}

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

func GetAll[T interface{}](collection string, filter bson.D, k int64) ([]T, error) {
	var results []T

	cursor, err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		Find(context.TODO(), filter, options.Find().SetLimit(k))
	if err != nil {
		return results, err
	}

	err = cursor.All(context.TODO(), &results)
	return results, err
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

func InitIndices() {
	userIndexes := mongoClient.Database("GoCharge").
		Collection("Users").
		Indexes()
	userIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{"username", 1}},
		Options: options.Index().SetUnique(true),
	})
	userIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{"email", 1}},
		Options: options.Index().SetUnique(true),
	})

	stationIndexes := mongoClient.Database("GoCharge").
		Collection("Stations").
		Indexes()
	stationIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{"coordinates", "2dsphere"}},
		Options: options.Index().SetUnique(true),
	})
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

	mongoClient = client

	InitIndices()
}
