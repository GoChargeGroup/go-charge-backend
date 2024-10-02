package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// NOTE: validating requests is done by gocharge admins manually (just change it in the DB).
func HandleStationRequest(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)

	if user_claim.Role != OWNER_ROLE {
		c.JSON(http.StatusUnauthorized, "Only owner accounts are allowed to request charging stations")
		return
	}

	station_data, err := ReadBodyToStruct[NewStationInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	station_id, err := CreateStation(user_id, station_data.Name, station_data.Description, station_data.Coordinates)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A station with this location must already exist")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	station, err := GetStation(bson.D{{"_id", station_id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// note: this should be parallelized, but i cant really bother.
	chargers := []Charger{}
	for _, charger := range station_data.Chargers {
		charger_id, err := CreateCharger(station_id, charger.Name, charger.Description, charger.KWhTypesId, charger.ChargerTypesId, charger.Price)
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, "A station with this location must already exist")
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		charger, err := GetCharger(bson.D{{"_id", charger_id}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		chargers = append(chargers, charger)
	}

	new_station_out := NewStationOutput{station, chargers}
	c.JSON(http.StatusOK, new_station_out)
}

// Get k-closest stations to a location
func HandleClosestStations(c *gin.Context) {
	query_data, err := ReadBodyToStruct[FindStationsInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	mongoDBHQ := bson.D{
		{"type", "Point"},
		{"coordinates", query_data.Coordinates},
	}
	stations, err := GetStations(bson.D{
		{"coordinates", bson.D{
			{"$near", bson.D{
				{"$geometry", mongoDBHQ},
				{"$maxDistance", query_data.Radius},
			}},
		}},
	}, query_data.K)

	result := FindStationsOutput{stations}
	c.JSON(http.StatusOK, result)
}
