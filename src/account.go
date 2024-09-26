package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleSignup(c *gin.Context) {
	var username, password, email, role string
	err := ReadQueryParams(c,
		QPPair{"username", &username},
		QPPair{"password", &password},
		QPPair{"email", &email},
		QPPair{"role", &role})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	_, err = CreateUser(username, password, email, role)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A user with this username already exists.")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"username", username}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	GenAndSetJWT(c, user.ID)
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
		{"password", password}})
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, "User not found.")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	GenAndSetJWT(c, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleEditAccount(c *gin.Context) {
	var username, password, email, role string
	err := ReadQueryParams(c,
		QPPair{"username", &username},
		QPPair{"password", &password},
		QPPair{"email", &email},
		QPPair{"role", &role})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	_, err = CreateUser(username, password, email, role)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A user with this username already exists.")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"username", username}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleDeleteAccount(c *gin.Context) {
	var username, password, email, role string
	err := ReadQueryParams(c,
		QPPair{"username", &username},
		QPPair{"password", &password},
		QPPair{"email", &email},
		QPPair{"role", &role})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	_, err = CreateUser(username, password, email, role)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A user with this username already exists.")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"username", username}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandlePasswordReset(c *gin.Context) {
	c.JSON(http.StatusOK, 0)
}
