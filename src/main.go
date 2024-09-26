package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	InitMongoDb()

	r := gin.Default()
	// define the routes
	r.GET("/signup", HandleSignup)
	r.GET("/login", HandleLogin)
	r.GET("/password-reset", HandlePasswordReset)
	err := r.Run(":8083")
	if err != nil {
		log.Fatalf("impossible to start server: %s", err)
	}

	// defer func() {
	// 	if err := mongoClient.Disconnect(context.TODO()); err != nil {
	// 		panic(err)
	// 	}
	// }()
}
