package main

import (
	"net/http"
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
	body_data, err := ReadBodyToStruct[SignupInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if body_data.Role != USER_ROLE && body_data.Role != OWNER_ROLE {
		c.JSON(http.StatusBadRequest, "Role must be 'user' or 'owner'")
		return
	}

	new_user := NewUser{
		Username:                body_data.Username,
		Password:                body_data.Password,
		Email:                   body_data.Email,
		Role:                    body_data.Role,
		PhotoURL:                "",
		FavoriteStationIDs:      []string{},
		SecurityQuestionAnswers: body_data.SecurityQuestionAnswers,
	}
	user_id, err := CreateOne(USER_COLL, new_user)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A user with this username or email already exists")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"_id", user_id}})
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
		c.JSON(http.StatusUnauthorized, "Incorrect username or password")
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

func HandleEditEmail(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// read body
	body_data, err := ReadBodyToStruct[ResetEmailInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if body_data.NewEmail == user_claim.Email {
		c.JSON(http.StatusConflict, "Cannot use the same email.")
		return
	}

	// update mongodb
	err = UpdateUser(
		bson.D{
			{"_id", user_id},
			{"security_question_answers", body_data.SecurityQuestionAnswers}, // ensure the security questions are correct
		},
		bson.D{
			{"$set", bson.D{
				{"email", body_data.NewEmail},
			}},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// send email to old email
	subject := "Email Update Notice"
	title := "GoCharge Email Update"
	action := "update your account's email to " + body_data.NewEmail
	msg := FormMessageBody(user_claim.Username, "", action, title)
	err = SendEmail(user_claim.Email, msg, subject) // user claim contains the old email, which is what we want.
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// reset jwt header since email or username changed.
	user, err := GetUser(bson.D{{"_id", user_id}})
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

var otp_manager = OTPManager{id_map: map[string]OTPData{}}

func HandleEditUsername(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)

	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// read body
	var otp, new_username string
	err = ReadBody(c, QPPair{"otp", &otp}, QPPair{"new_username", &new_username})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// try otp
	err = otp_manager.TryOTP(user_claim.ID, otp)
	if err != nil {
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	// perform update
	err = UpdateUser(
		bson.D{{"_id", user_id}},
		bson.D{
			{"$set", bson.D{
				{"username", new_username},
			}},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// get new user doc and update jwt
	user, err := GetUser(bson.D{{"_id", user_id}})
	if err != nil || user.Email == "" {
		c.JSON(http.StatusOK, "")
		return
	}
	err = GenAndSetJWT(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleEditUsernameRequest(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)

	// gen otp
	otp_data := otp_manager.GenOTP(user_claim.ID)

	// send email
	subject := "Edit Username Request"
	title := "GoCharge Edit Username"
	action := "edit your GoCharge username"
	msg := FormMessageBody(user_claim.Username, otp_data.otp, action, title)
	err := SendEmail(user_claim.Email, msg, subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, OTPResponse{otp_data.expiration})
}

func HandleDeleteAccount(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)

	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// read body
	var otp string
	err = ReadBody(c, QPPair{"otp", &otp})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// try otp
	err = otp_manager.TryOTP(user_claim.ID, otp)
	if err != nil {
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	// delete user
	err = DeleteUser(bson.D{{"_id", user_id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}

func HandleDeleteAccountRequest(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	otp_data := otp_manager.GenOTP(user_id.String())

	title := "GoCharge Delete Account"
	action := "delete your GoCharge account"
	subject := "Delete Account Request"
	msg := FormMessageBody(user_claim.Username, otp_data.otp, action, title)
	err = SendEmail(user_claim.Email, msg, subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, OTPResponse{otp_data.expiration})
}

func HandlePasswordReset(c *gin.Context) {
	var otp, email, password string
	err := ReadBody(c, QPPair{"otp", &otp}, QPPair{"email", &email}, QPPair{"password", &password})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"email", email}})
	if err != nil {
		c.JSON(http.StatusNotFound, "It appears there no longer exist an account using this email.")
		return
	}
	if user.Role == ADMIN_ROLE {
		c.JSON(http.StatusUnauthorized, "If you are an admin, please contact GoCharge dev team to reset your password.")
		return
	}

	err = otp_manager.TryOTP(user.ID, otp)
	if err != nil {
		c.JSON(http.StatusUnauthorized, err.Error())
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

	otp_data := otp_manager.GenOTP(user.ID)

	title := "GoCharge Password Reset"
	action := "reset your GoCharge password"
	subject := "Password Reset Request"
	msg := FormMessageBody(user.Username, otp_data.otp, action, title)
	err = SendEmail(user.Email, msg, subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, OTPResponse{otp_data.expiration})
}
