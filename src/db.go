package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

const SESSION_COLL = "Sessions"
const STATION_COLL = "Stations"
const USER_COLL = "Users"
const CHARGER_COLL = "Chargers"
const REVIEW_COLL = "Reviews"

// STATION WRAPPER FUNCTIONS

func CreateStation(new_station Station) (primitive.ObjectID, error) {
	return CreateOne(STATION_COLL, new_station)
}

func GetStation(filter bson.D) (Station, error) {
	return GetOne[Station](STATION_COLL, filter)
}

func GetStations(filter bson.D, max_results int64) ([]Station, error) {
	return GetAll[Station](STATION_COLL, filter, max_results)
}

// STATION WRAPPER FUNCTIONS

func CreateCharger(new_charger Charger) (primitive.ObjectID, error) {
	return CreateOne(CHARGER_COLL, new_charger)
}

func GetCharger(filter bson.D) (Charger, error) {
	return GetOne[Charger](CHARGER_COLL, filter)
}

// USER WRAPPER FUNCTIONS

func GetUser(filter bson.D) (User, error) {
	return GetOne[User](USER_COLL, filter)
}

func UpdateUser(filter bson.D, update bson.D) error {
	return UpdateOne(USER_COLL, filter, update)
}

func DeleteUser(filter bson.D) error {
	return DeleteOne(USER_COLL, filter)
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

func GetAll[T interface{}](collection string, filter bson.D, max_results int64) ([]T, error) {
	results := []T{}

	cursor, err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		Find(context.TODO(), filter, options.Find().SetLimit(max_results))
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

func UpdateOne(collection string, filter interface{}, update interface{}) error {
	res, err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("No record found to update")
	}
	if res.ModifiedCount == 0 {
		return errors.New("No records modified")
	}
	return err
}

func Aggregate[T interface{}](collection string, pipeline bson.A) ([]T, error) {
	results := []T{}

	cursor, err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		Aggregate(context.TODO(), pipeline)
	if err != nil {
		return results, err
	}

	err = cursor.All(context.TODO(), &results)
	return results, err
}

func DeleteOne(collection string, filter interface{}) error {
	_, err := mongoClient.
		Database("GoCharge").
		Collection(collection).
		DeleteOne(context.TODO(), filter)
	return err
}

func InitIndices() {
	userIndexes := mongoClient.Database("GoCharge").
		Collection(USER_COLL).
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
		Collection(STATION_COLL).
		Indexes()
	stationIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{"coordinates", "2dsphere"}},
		Options: options.Index().SetUnique(true),
	})
	stationIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{"owner_id", 1}},
	})

	chargerIndexes := mongoClient.Database("GoCharge").
		Collection(CHARGER_COLL).
		Indexes()
	chargerIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{"station_id", 1}},
	})
	chargerIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{"station_id", 1}, {"name", 1}},
		Options: options.Index().SetUnique(true),
	})

	sessionIndexes := mongoClient.Database("GoCharge").
		Collection(SESSION_COLL).
		Indexes()
	sessionIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{"user_id", 1}},
	})
	sessionIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{"charger_id", 1}},
	})
	sessionIndexes.CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{"end_timestamp", 1}},
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
