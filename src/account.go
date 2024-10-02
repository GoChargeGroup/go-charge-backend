package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const USER_ROLE = "user"
const OWNER_ROLE = "owner"

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
	if role != USER_ROLE && role != OWNER_ROLE {
		c.JSON(http.StatusBadRequest, "Role must be 'user' or 'owner'")
		return
	}

	userID, err := CreateUser(username, password, email, role)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A user with this username or email already exists")
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

	err = GenAndSetJWT(c, user)
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
		c.JSON(http.StatusNotFound, "Incorrect username or password")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = GenAndSetJWT(c, user)
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

	// reset jwt header since username and email might have changed.
	err = GenAndSetJWT(c, user)
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
	userClaim := c.MustGet(MW_USER_KEY).(UserClaim)

	var password string
	err := ReadBody(c, QPPair{"password", &password})
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
				{"password", password},
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

func HandlePasswordResetRequest(c *gin.Context) {
	var email string
	err := ReadBody(c, QPPair{"email", &email})
	if err != nil {
		c.JSON(http.StatusNotFound, "Email not provided")
		return
	}

	user, err := GetUser(bson.D{{"email", email}})
	if err != nil || user.Email == "" {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	msg, err := GetResetPasswordMessageBody(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = SendEmail(user, msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, msg)
}
