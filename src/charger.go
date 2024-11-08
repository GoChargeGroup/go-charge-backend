package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleAddCharger(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	body_data, err := ReadBodyToStruct[AddChargerInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// ensure owner is authorized.
	_, err = GetStation(bson.D{
		{"_id", body_data.StationID},
		{"owner_id", user_id},
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, "Owner does not own a station of this id")
		return
	}

	// create new charger
	new_charger := Charger{
		ID:             primitive.NewObjectID(),
		StationID:      body_data.StationID,
		Name:           body_data.Name,
		Description:    body_data.Description,
		KWhTypesId:     body_data.KWhTypesId,
		ChargerTypesId: body_data.ChargerTypesId,
		Status:         "working",
		Price:          body_data.Price,
		TotalPayments:  0,
	}
	charger_id, err := CreateCharger(new_charger)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A station with this location already exists")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// get and return new charger doc
	charger, err := GetCharger(bson.D{{"_id", charger_id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, charger)
}

func HandleEditCharger(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	body_data, err := ReadBodyToStruct[EditChargerInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// ensure owner is authorized.
	_, err = GetStation(bson.D{
		{"_id", body_data.StationID},
		{"owner_id", user_id},
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, "Owner does not own a station of this id")
		return
	}

	// edit charger
	err = UpdateOne(
		CHARGER_COLL,
		bson.D{
			{"_id", body_data.ID},
		},
		bson.D{
			{"$set", bson.D{
				{"name", body_data.Name},
				{"description", body_data.Description},
				{"kWh_types_id", body_data.KWhTypesId},
				{"charger_types_id", body_data.ChargerTypesId},
				{"price", body_data.Price},
				{"status", body_data.Status},
			}},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// get and return updated charger doc
	charger, err := GetCharger(bson.D{{"_id", body_data.ID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, charger)
}
