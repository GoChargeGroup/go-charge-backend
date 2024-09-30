package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Pair[T any, K any] struct {
	_0 T
	_1 K
}

type QPPair = Pair[string, *string]

func ReadQueryParams(c *gin.Context, key_vars ...QPPair) error {
	for _, key_var := range key_vars {
		value := c.Query(key_var._0)
		if value == "" {
			return errors.New("no " + key_var._0 + " provided")
		}
		*key_var._1 = value
	}
	return nil
}

func ReadBody(c *gin.Context, key_vars ...QPPair) error {
	bodyJson, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return errors.New("invalid request body")
	}

	jsonMap := map[string]interface{}{}
	err = json.Unmarshal([]byte(bodyJson), &jsonMap)
	if err != nil {
		return errors.New("invalid request body")
	}

	for _, key_var := range key_vars {
		value, ok := jsonMap[key_var._0]
		if !ok {
			return errors.New("no " + key_var._0 + " provided")
		}
		*key_var._1 = fmt.Sprint(value)
	}
	return nil
}

func Unpack(s []string, vars ...*string) {
	for i, str := range s {
		*vars[i] = str
	}
}

func ID2Str(id any) string {
	return id.(primitive.ObjectID).Hex()
}

const JWT_KEY = "hi mom"
const JWT_PREFIX = "Bearer "
const JWT_HEADER = "Authorization"
const MW_USER_KEY = "user"

func GenAndSetJWT(c *gin.Context, user User) error {
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
		return err
	}

	c.Header(JWT_HEADER, JWT_PREFIX+jwt_token_str)
	return nil
}

func AuthMiddleware(c *gin.Context) {
	auth_str := c.GetHeader(JWT_HEADER)
	jwt_token_str := strings.Replace(auth_str, JWT_PREFIX, "", 1)

	token, err := jwt.ParseWithClaims(jwt_token_str, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_KEY), nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if !token.Valid {
		c.JSON(http.StatusUnauthorized, "Invalid token.")
		return
	}

	claims := token.Claims.(*JWTClaims)
	c.Set(MW_USER_KEY, claims.User)
	c.Next()
}
