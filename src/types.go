package main

import (
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	User UserClaim `json:"user"`
}

type UserClaim struct {
	ID       string `json:"_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type NewUser struct {
	FavoriteStationIDs []string `json:"favorite_station_ids" bson:"favorite_station_ids"`
	Username           string   `json:"username" bson:"username"`
	Password           string   `json:"password" bson:"password"`
	Email              string   `json:"email" bson:"email"`
	Role               string   `json:"role" bson:"role"`
	PhotoURL           string   `json:"photo_url" bson:"photo_url"`
}

type User struct {
	ID                 string   `json:"_id" bson:"_id"`
	FavoriteStationIDs []string `json:"favorite_station_ids" bson:"favorite_station_ids"`
	Username           string   `json:"username" bson:"username"`
	Email              string   `json:"email" bson:"email"`
	Role               string   `json:"role" bson:"role"`
	PhotoURL           string   `json:"photo_url" bson:"photo_url"`
}

type FindStationsInput struct {
	Statuses     []string   `json:"statuses"`
	PowerOutputs []string   `json:"power_outputs"`
	PlugTypes    []string   `json:"plug_types"`
	MaxPrice     float64    `json:"max_price"`
	MaxRadius    float64    `json:"max_radius"`
	MaxResults   int64      `json:"max_results"`
	Coordinates  [2]float64 `json:"coordinates"`
}

type UnapprovedStationsOutput struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	OwnerID          primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	PictureURLs      []string           `json:"picture_urls" bson:"picture_urls"`
	Name             string             `json:"name" bson:"name"`
	Description      string             `json:"description" bson:"description"`
	Coordinates      [2]float64         `json:"coordinates" bson:"coordinates"`
	Address          string             `json:"address" bson:"address"`
	IsPublic         bool               `json:"is_public" bson:"is_public"`
	OperationalHours [7][2]int64        `json:"operational_hours" bson:"operational_hours"` // format: [days of week][start, end]sec_since_start_of_UNIX_day
	Chargers         []Charger          `json:"chargers" bson:"chargers"`
}

type FindStationsOutput struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	OwnerID          primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	PictureURLs      []string           `json:"picture_urls" bson:"picture_urls"`
	Name             string             `json:"name" bson:"name"`
	Description      string             `json:"description" bson:"description"`
	Coordinates      [2]float64         `json:"coordinates" bson:"coordinates"`
	Address          string             `json:"address" bson:"address"`
	IsPublic         bool               `json:"is_public" bson:"is_public"`
	OperationalHours [7][2]int64        `json:"operational_hours" bson:"operational_hours"` // format: [days of week][start, end]sec_since_start_of_UNIX_day
	Chargers         []Charger          `json:"chargers" bson:"chargers"`
	Distance         float64            `json:"distance" bson:"distance"`
}

type ApprovedStationInput struct {
	StationID string `json:"station_id" bson:"station_id"`
}

type FavoriteStationInput struct {
	StationID string `json:"station_id" bson:"station_id"`
}

type NewStationInput struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Coordinates      [2]float64        `json:"coordinates"`
	Address          string            `json:"address"`
	OperationalHours [7][2]int64       `json:"operational_hours"`
	Chargers         []NewChargerInput `json:"chargers"`
}

type NewStationOutput struct {
	Station  Station   `json:"station"`
	Chargers []Charger `json:"chargers"`
}

type Station struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	OwnerID          primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	PictureURLs      []string           `json:"picture_urls" bson:"picture_urls"`
	Name             string             `json:"name" bson:"name"`
	Description      string             `json:"description" bson:"description"`
	Coordinates      [2]float64         `json:"coordinates" bson:"coordinates"`
	Address          string             `json:"address" bson:"address"`
	IsPublic         bool               `json:"is_public" bson:"is_public"`
	OperationalHours [7][2]int64        `json:"operational_hours" bson:"operational_hours"` // format: [days of week][start, end]sec_since_start_of_UNIX_day
}

type NewChargerInput struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	KWhTypesId     string  `json:"kWh_types_id"`
	ChargerTypesId string  `json:"charger_types_id"`
	Price          float64 `json:"price"`
}

type Charger struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	StationID      primitive.ObjectID `json:"station_id" bson:"station_id"`
	Name           string             `json:"name" bson:"name"`
	Description    string             `json:"description" bson:"description"`
	KWhTypesId     string             `json:"kWh_types_id" bson:"kWh_types_id"`
	ChargerTypesId string             `json:"charger_types_id" bson:"charger_types_id"`
	Status         string             `json:"status" bson:"status"`
	Price          float64            `json:"price" bson:"price"`
	TotalPayments  float64            `json:"total_payments" bson:"total_payments"`
}

type NewSessionInput struct {
	ChargerID primitive.ObjectID `json:"charger_id"`
}

type EndSessionInput struct {
	ID            primitive.ObjectID `json:"_id"`
	PaymentAmount float64            `json:"payment_amount"`
	PowerUsed     float64            `json:"power_used"`
}

type Session struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	UserID         primitive.ObjectID `json:"user_id" bson:"user_id"`
	ChargerID      primitive.ObjectID `json:"charger_id" bson:"charger_id"`
	StartTimestamp int64              `json:"start_timestamp" bson:"start_timestamp"` // start unix timestamp
	EndTimestamp   int64              `json:"end_timestamp" bson:"end_timestamp"`     // end unix timestamp
	PaymentAmount  float64            `json:"payment_amount" bson:"payment_amount"`
	PowerUsed      float64            `json:"power_used" bson:"power_used"`
}

// type ChargerReview struct {
// 	ID         primitive.ObjectID `json:"_id"`
// 	UserID     string             `json:"user_id"`
// 	ChargerID  string             `json:"charger_id"`
// 	PhotoURLs  []string           `json:"photo_urls"`
// 	Rating     int                `json:"rating"`
// 	Commentary string             `json:"commentary"`
// }
