package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleStartSession(c *gin.Context) {
	start_session_data, err := ReadBodyToStruct[NewSessionInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	session := Session{
		ID:             primitive.NewObjectID(),
		UserID:         user_id,
		ChargerID:      start_session_data.ChargerID,
		StartTimestamp: time.Now().UnixMicro(),
		EndTimestamp:   0,
		PaymentAmount:  0,
		PowerUsed:      0,
	}
	object_id, err := CreateOne(SESSION_COLL, session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	session, err = GetOne[Session](SESSION_COLL, bson.D{{"_id", object_id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, session)
}

func HandleEndSession(c *gin.Context) {
	end_session_data, err := ReadBodyToStruct[EndSessionInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	filter := bson.D{
		{"_id", end_session_data.ID},
		{"user_id", user_id},
	}
	err = UpdateOne(SESSION_COLL, filter, end_session_data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	session, err := GetOne[Session](SESSION_COLL, bson.D{{"_id", end_session_data.ID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, session)
}