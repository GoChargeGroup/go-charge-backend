package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const USER_ROLE = "user"
const OWNER_ROLE = "owner"
const ADMIN_ROLE = "admin"

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

func HandleLogout(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// user must end session first.
	_, err = GetOne[Session](SESSION_COLL, bson.D{
		{"user_id", user_id},
		{"end_timestamp", 0},
	})
	if err == nil {
		c.JSON(http.StatusConflict, "Cannot log out until the current session has been finished.")
		return
	}

	// log time user last logged out.
	err = UpdateUser(
		bson.D{
			{"_id", user_id},
		},
		bson.D{
			{"$set", bson.D{
				{"last_logout_timestamp", time.Now().Unix()},
			}},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// clear auth cookie
	c.SetCookie("Authorization", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

var edit_account_otp_id_map = map[string]string{}

func HandleEditAccount(c *gin.Context) {
	userClaim := c.MustGet(MW_USER_KEY).(UserClaim)

	var otp, username, email string
	err := ReadBody(c,
		QPPair{"username", &username},
		QPPair{"email", &email},
		QPPair{"otp", &otp},
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user_id_str, ok := edit_account_otp_id_map[otp]
	if !ok || user_id_str != userClaim.ID {
		c.JSON(http.StatusUnauthorized, "Invalid OTP.")
		return
	}
	delete(edit_account_otp_id_map, otp)

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

func HandleEditAccountRequest(c *gin.Context) {
	userClaim := c.MustGet(MW_USER_KEY).(UserClaim)

	user, err := GetUser(bson.D{{"email", userClaim.Email}})
	if err != nil || user.Email == "" {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	otp := strconv.Itoa(rand.Int() % 1000)
	edit_account_otp_id_map[otp] = user.ID

	msg, err := GetResetPasswordMessageBody(user, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = SendEmail(user, msg, "Edit Account Request")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}

var delete_account_otp_id_map = map[string]string{}

func HandleDeleteAccount(c *gin.Context) {
	userClaim := c.MustGet(MW_USER_KEY).(UserClaim)

	userID, err := primitive.ObjectIDFromHex(userClaim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var otp string
	err = ReadBody(c, QPPair{"otp", &otp})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user_id_str, ok := delete_account_otp_id_map[otp]
	if !ok || user_id_str != userClaim.ID {
		c.JSON(http.StatusUnauthorized, "Invalid OTP.")
		return
	}
	delete(delete_account_otp_id_map, otp)

	err = DeleteUser(bson.D{{"_id", userID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}

func HandleDeleteAccountRequest(c *gin.Context) {
	userClaim := c.MustGet(MW_USER_KEY).(UserClaim)

	user, err := GetUser(bson.D{{"email", userClaim.Email}})
	if err != nil || user.Email == "" {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	otp := strconv.Itoa(rand.Int() % 1000)
	delete_account_otp_id_map[otp] = user.ID

	msg, err := GetDeleteAccountMessageBody(user, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = SendEmail(user, msg, "Delete Account Request")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}

var reset_pwd_otp_id_map = map[string]string{}

func HandlePasswordReset(c *gin.Context) {
	var otp, email, password string
	err := ReadBody(c, QPPair{"otp", &otp}, QPPair{"email", &email}, QPPair{"password", &password})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user_email_str, ok := reset_pwd_otp_id_map[otp]
	if !ok || user_email_str != email {
		c.JSON(http.StatusUnauthorized, "Invalid OTP.")
		return
	}
	delete(reset_pwd_otp_id_map, otp)

	user, err := GetUser(bson.D{{"email", email}})
	if err != nil {
		c.JSON(http.StatusNotFound, "It appears there no longer exist an account using this email.")
		return
	}
	if user.Role == ADMIN_ROLE {
		c.JSON(http.StatusUnauthorized, "If you are an admin, please contact GoCharge dev team to reset your password.")
		return
	}

	err = UpdateUser(
		bson.D{
			{"email", email},
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

	err = GenAndSetJWT(c, user)
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
		c.JSON(http.StatusOK, "")
		return
	}

	otp := strconv.Itoa(rand.Int() % 1000)
	reset_pwd_otp_id_map[otp] = email

	msg, err := GetResetPasswordMessageBody(user, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = SendEmail(user, msg, "Password Reset Request")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}
