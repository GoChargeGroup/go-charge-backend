package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	InitPasswordResetTemplate()
	InitGmailService()
	InitMongoDb()

	router := gin.Default()
	router.GET("/signup", HandleSignup)
	router.GET("/login", HandleLogin)
	router.POST("/password-reset-request", HandlePasswordResetRequest)

	user_router := router.Group("/user")
	user_router.Use(AuthMiddleware)
	user_router.POST("/edit-account", HandleEditAccount)
	user_router.POST("/delete-account", HandleDeleteAccount)
	user_router.POST("/password-reset", HandlePasswordReset)

	err := router.Run(":8083")
	if err != nil {
		log.Fatalf("impossible to start server: %s", err)
	}

	// defer func() {
	// 	if err := mongoClient.Disconnect(context.TODO()); err != nil {
	// 		panic(err)
	// 	}
	// }()
}
