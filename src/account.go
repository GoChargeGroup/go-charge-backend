package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleSignup(c *gin.Context) {
	var username, password, email, role string
	err := ReadQueryParams(c,
		QPPair{"username", &username},
		QPPair{"password", &password},
		QPPair{"email", &email},
		QPPair{"role", &role},
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	userID, err := CreateUser(username, password, email, role)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A user with this username already exists.")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"_id", userID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	GenAndSetJWT(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleLogin(c *gin.Context) {
	var username, password string
	err := ReadQueryParams(c,
		QPPair{"username", &username},
		QPPair{"password", &password})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := GetUser(bson.D{
		{"username", username},
		{"password", password},
	})
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, "User not found.")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	GenAndSetJWT(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleEditAccount(c *gin.Context) {
	userClaim := c.MustGet(MW_USER_KEY).(UserClaim)

	var username, email string
	err := ReadBody(c,
		QPPair{"username", &username},
		QPPair{"email", &email},
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	userID, err := primitive.ObjectIDFromHex(userClaim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = UpdateUser(
		bson.D{
			{"_id", userID},
		},
		bson.D{
			{"$set", bson.D{
				{"username", username},
				{"email", email},
			}},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"_id", userID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleDeleteAccount(c *gin.Context) {
	userClaim := c.MustGet(MW_USER_KEY).(UserClaim)

	userID, err := primitive.ObjectIDFromHex(userClaim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = DeleteUser(bson.D{{"_id", userID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}

func HandlePasswordReset(c *gin.Context) {
	c.JSON(http.StatusOK, "")
}
