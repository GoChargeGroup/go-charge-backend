package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleReviewStation(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	body_data, err := ReadBodyToStruct[NewReviewInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// insert new review
	new_review := NewReview{
		UserID:     user_id,
		StationID:  body_data.StationID,
		ChargerID:  body_data.ChargerID,
		PhotoURLs:  body_data.PhotoURLs,
		Rating:     body_data.Rating,
		Commentary: body_data.Commentary,
		CreatedAt:  time.Now(),
	}
	review_id, err := CreateOne(REVIEW_COLL, new_review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// update aggregate info in station
	err = UpdateOne(
		STATION_COLL,
		bson.D{{"_id", body_data.StationID}},
		bson.D{{"$inc", bson.D{
			{"review_count", 1},
			{"review_score", new_review.Rating},
		}}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// get and send new review doc
	review, err := GetOne[Review](REVIEW_COLL, bson.D{{"_id", review_id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, review)
}

func HandleGetStationReviews(c *gin.Context) {
	body_data, err := ReadBodyToStruct[GetStationReviewsInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	reviews, err := GetAll[Review](REVIEW_COLL, bson.D{{"station_id", body_data.StationID}}, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func HandleGetMyReviews(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	reviews, err := GetAll[Review](REVIEW_COLL, bson.D{{"user_id", user_id}}, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, reviews)
}
