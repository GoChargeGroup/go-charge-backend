package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	// search for ongoing sessions with the current charger, or the current user.
	conflicting_session, err := GetOne[Session](SESSION_COLL, bson.D{
		{"$or",
			bson.A{
				bson.D{{"user_id", user_id}},
				bson.D{{"charger_id", start_session_data.ChargerID}},
			},
		},
		{"end_timestamp", 0}, // end_timestamp of 0 means not done.
	})
	if err == nil {
		same_user := conflicting_session.UserID == user_id
		same_charger := conflicting_session.ChargerID == start_session_data.ChargerID
		if same_user && same_charger {
			c.JSON(http.StatusInternalServerError, "This user already has a session open with this charger")
			return
		}
		if same_user {
			c.JSON(http.StatusInternalServerError, "This user already has an existing session open")
			return
		}
		if same_charger {
			c.JSON(http.StatusInternalServerError, "This charger already has an existing session open")
			return
		}
		c.JSON(http.StatusInternalServerError, "Impossible case reached")
		return
	}

	session := Session{
		ID:             primitive.NewObjectID(),
		UserID:         user_id,
		ChargerID:      start_session_data.ChargerID,
		StartTimestamp: time.Now().Unix(),
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
		{"end_timestamp", 0},
	}
	update := bson.D{
		{"payment_amount", end_session_data.PaymentAmount},
		{"power_used", end_session_data.PowerUsed},
		{"end_timestamp", time.Now().Unix()},
	}
	err = UpdateOne(SESSION_COLL, filter, update)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, "No such session found.")
		return
	}
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
