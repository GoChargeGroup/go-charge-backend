package main

import (
	"log"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(router *gin.Engine) {
	user_router := router.Group("/user")
	user_router.Use(BuildAuthMiddleware(USER_ROLE))

	// account routes
	user_router.POST("/edit-account", HandleEditAccount)
	user_router.POST("/edit-account-request", HandleEditAccountRequest)
	user_router.POST("/delete-account", HandleDeleteAccount)
	user_router.POST("/delete-account-request", HandleDeleteAccountRequest)
	user_router.POST("/logout", HandleLogout)

	// station routes
	user_router.POST("/closest-stations", HandleClosestStations)
	user_router.POST("/favorite-station", HandleFavoriteStation)
	user_router.POST("/unfavorite-station", HandleUnfavoriteStation)

	// session routes
	user_router.POST("/start-session", HandleStartSession)
	user_router.POST("/end-session", HandleEndSession)
}

func InitOwnerRouter(router *gin.Engine) {
	owner_router := router.Group("/owner")
	owner_router.Use(BuildAuthMiddleware(OWNER_ROLE))

	// account routes
	owner_router.POST("/edit-account", HandleEditAccount)
	owner_router.POST("/edit-account-request", HandleEditAccountRequest)
	owner_router.POST("/delete-account", HandleDeleteAccount)
	owner_router.POST("/delete-account-request", HandleDeleteAccountRequest)
	owner_router.POST("/logout", HandleLogout)

	// station routes
	owner_router.POST("/request-station", HandleStationRequest)
}

func InitAdminRouter(router *gin.Engine) {
	admin_router := router.Group("/admin")
	admin_router.Use(BuildAuthMiddleware(ADMIN_ROLE))

	// account routes
	admin_router.POST("/logout", HandleLogout)

	// station routes
	admin_router.POST("/approve-station", HandleStationRequestApproval)
	admin_router.POST("/unapproved-stations", HandleUnapprovedStations)
}

var wg sync.WaitGroup

func RunPublicVersion() {
	router := gin.Default()

	router.GET("/signup", HandleSignup)
	router.GET("/login", HandleLogin)
	router.POST("/password-reset-request", HandlePasswordResetRequest)
	router.POST("/password-reset", HandlePasswordReset)

	InitUserRouter(router)
	InitOwnerRouter(router)

	err := router.Run(":8083")
	if err != nil {
		log.Fatalf("impossible to start server: %s", err)
	}

	defer wg.Done()
}

func RunPrivateVersion() {
	router := gin.Default()

	router.GET("/login", HandleLogin)

	InitAdminRouter(router)

	err := router.Run(":8084")
	if err != nil {
		log.Fatalf("impossible to start server: %s", err)
	}

	defer wg.Done()
}

func main() {
	InitPasswordResetTemplate()
	InitGmailService()
	InitMongoDb()

	run_public_version := len(os.Args) < 2 || os.Args[1] == "public"
	if run_public_version {
		wg.Add(1)
		go RunPublicVersion()
	}

	run_private_version := len(os.Args) < 2 || os.Args[1] == "private"
	if run_private_version {
		wg.Add(1)
		go RunPrivateVersion()
	}

	wg.Wait()
}
