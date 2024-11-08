package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const JWT_KEY = "hi mom"
const JWT_PREFIX = "Bearer "
const JWT_HEADER = "Authorization"
const MW_USER_KEY = "user"

func GenJWT(user User) (string, error) {
	claims := JWTClaims{
		jwt.RegisteredClaims{},
		UserClaim{
			user.ID,
			user.Username,
			user.Email,
			user.Role,
		},
	}
	jwt_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwt_token_str, err := jwt_token.SignedString([]byte(JWT_KEY))
	if err != nil {
		return "", err
	}

	return jwt_token_str, nil
}

func GenAndSetJWT(c *gin.Context, user User) error {
	jwt_token_str, err := GenJWT(user)
	if err != nil {
		return err
	}

	c.Header(JWT_HEADER, JWT_PREFIX+jwt_token_str)
	return nil
}

func BuildAuthMiddleware(required_role string) func(*gin.Context) {
	return func(c *gin.Context) {
		AuthMiddleware(c, required_role)
		return
	}
}

func AuthMiddleware(c *gin.Context, required_role string) {
	auth_header := c.GetHeader(JWT_HEADER)
	if auth_header == "" {
		c.JSON(http.StatusUnauthorized, "Authorization header is missing")
		c.Abort()
		return
	}

	token_str := strings.TrimPrefix(auth_header, JWT_PREFIX)
	if token_str == auth_header {
		c.JSON(http.StatusUnauthorized, "Invalid token format")
		c.Abort()
		return
	}

	token, err := jwt.ParseWithClaims(token_str, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_KEY), nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Failed to parse token")
		c.Abort()
		return
	}
	if !token.Valid {
		c.JSON(http.StatusUnauthorized, "Invalid token")
		c.Abort()
		return
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || claims.User.ID == "" {
		c.JSON(http.StatusUnauthorized, "Invalid token claims")
		c.Abort()
		return
	}
	if claims.User.Role != required_role {
		msg := claims.User.Role + " cannot perform this action"
		c.JSON(http.StatusUnauthorized, msg)
		c.Abort()
		return
	}

	c.Set(MW_USER_KEY, claims.User)
	c.Next()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

type OTPData struct {
	otp        string
	expiration int64
}

type OTPManager struct {
	id_map map[string]OTPData
}

func (m OTPManager) GenOTP(user_id string) OTPData {
	otp := fmt.Sprintf("%05d", rand.Int()%100000)

	curr_time := time.Now().Unix()
	exp_time := curr_time + 30 // expires in 30 seconds

	otp_data := OTPData{otp, exp_time}
	m.id_map[user_id] = otp_data

	return otp_data
}

// func (m OTPManager) GenOTP(user_id string) (OTPData, error) {
// 	curr_time := time.Now().Unix()

// 	otp_data, ok := m.id_map[user_id]
// 	if ok && curr_time < otp_data.expiration {
// 		return OTPData{}, errors.New("Previous OTP has not expired yet")
// 	}

// 	return m.regen_otp(user_id), nil
// }

func (m OTPManager) TryOTP(user_id string, otp_attempt string) error {
	curr_time := time.Now().Unix()

	otp_data, ok := m.id_map[user_id]
	if !ok {
		return errors.New("No OTP has been generated for this user")
	}

	delete(m.id_map, user_id) // no matter what, delete the OTP.

	if curr_time >= otp_data.expiration {
		return errors.New("OTP has expired")
	}
	if otp_data.otp != otp_attempt {
		return errors.New("Incorrect OTP")
	}
	return nil
}
