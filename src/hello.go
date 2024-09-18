package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserByID(id string) (User, error) {
	// TODO : lookup in db
	return User{
		ID:       id,
		Username: "Doe",
	}, nil
}

func GetUserHandler(c *gin.Context) {
	// retrieve User from db
	id := c.Query("id")
	User, err := GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible to retrieve User"})
		return
	}
	c.JSON(http.StatusOK, User)
}

func main() {
	initMongoDb()
	r := gin.Default()
	// define the routes
	r.GET("/user", GetUserHandler)
	err := r.Run(":8083")
	if err != nil {
		log.Fatalf("impossible to start server: %s", err)
	}
}
