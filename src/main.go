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
	user_router.POST("/edit-email", HandleEditEmail)
	user_router.POST("/edit-username", HandleEditUsername)
	user_router.POST("/edit-username-request", HandleEditUsernameRequest)
	user_router.POST("/delete-account", HandleDeleteAccount)
	user_router.POST("/delete-account-request", HandleDeleteAccountRequest)
	user_router.POST("/logout", HandleLogout)

	// station routes
	user_router.POST("/closest-stations", HandleClosestStations)
	user_router.POST("/favorite-station", HandleFavoriteStation)
	user_router.POST("/unfavorite-station", HandleUnfavoriteStation)
	user_router.POST("/station-and-chargers", HandleGetStationAndChargers)

	// session routes
	user_router.POST("/start-session", HandleStartSession)
	user_router.POST("/end-session", HandleEndSession)

	// review routes
	user_router.POST("/review-station", HandleReviewStation)
	user_router.POST("/station-reviews", HandleGetStationReviews)
	user_router.POST("/my-reviews", HandleGetMyReviews)
}

func InitOwnerRouter(router *gin.Engine) {
	owner_router := router.Group("/owner")
	owner_router.Use(BuildAuthMiddleware(OWNER_ROLE))

	// account routes
	owner_router.POST("/edit-email", HandleEditEmail)
	owner_router.POST("/edit-username", HandleEditUsername)
	owner_router.POST("/edit-username-request", HandleEditUsernameRequest)
	owner_router.POST("/delete-account", HandleDeleteAccount)
	owner_router.POST("/delete-account-request", HandleDeleteAccountRequest)
	owner_router.POST("/logout", HandleLogout)

	// station routes
	owner_router.POST("/request-station", HandleStationRequest)
	owner_router.GET("/get-user-chargers", HandleGetUserChargers)
	owner_router.POST("/station-and-chargers", HandleGetStationAndChargers)
	owner_router.POST("/edit-station", HandleEditStation)

	// charger routes
	owner_router.POST("/add-charger", HandleAddCharger)
	owner_router.POST("/edit-charger", HandleEditCharger)
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
	router.Use(CORSMiddleware())

	router.POST("/signup", HandleSignup)
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
	router.Use(CORSMiddleware())

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
