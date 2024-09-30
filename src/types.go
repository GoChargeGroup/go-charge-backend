package main

import "github.com/golang-jwt/jwt/v5"

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

type Station struct {
	ID          string     `json:"id"`
	OwnerID     string     `json:"owner_id"`
	PictureURLs []string   `json:"picture_urls"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Coords      [2]float32 `json:"coords"`
	IsPublic    bool       `json:"is_public"`
}

type Charger struct {
	ID             string  `json:"id"`
	StationID      string  `json:"station_id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	KWhTypesId     string  `json:"kWh_types_id"`
	ChargerTypesId string  `json:"charger_types_id"`
	Status         string  `json:"status"`
	Price          float32 `json:"price"`
	TotalPayments  float32 `json:"total_payments"`
}

type Log struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	ChargerID      string  `json:"charger_id"`
	StartTimestamp uint64  `json:"start_timestamp"`
	EndTimestamp   uint64  `json:"end_timestamp"`
	PaymentAmount  string  `json:"payment_amount"`
	PowerUsed      float32 `json:"power_used"`
}

type StationRequest struct {
	ID          string `json:"id"`
	OwnerID     string `json:"owner_id"`
	StationID   string `json:"station_id"`
	Description string `json:"description"`
}

type ChargerReview struct {
	ID         string   `json:"id"`
	UserID     string   `json:"user_id"`
	ChargerID  string   `json:"charger_id"`
	PhotoURLs  []string `json:"photo_urls"`
	Rating     int      `json:"rating"`
	Commentary string   `json:"commentary"`
}
