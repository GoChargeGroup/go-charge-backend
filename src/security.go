package main

import (
	"net/http"
	"strings"

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

func AuthMiddleware(c *gin.Context) {
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

	c.Set(MW_USER_KEY, claims.User)
	c.Next()
}
