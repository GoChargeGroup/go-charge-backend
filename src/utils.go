package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	for _, key_var := range key_vars {
		value, exists := c.Get(key_var._0)
		if !exists {
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

type Item struct {
	Data map[string]interface{} `bson:"data"`
}

func ToMongoDoc(val any) (any, error) {
	json_doc, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	item := Item{}
	err = json.Unmarshal(json_doc, &item.Data)
	if err != nil {
		return nil, err
	}

	return item.Data, nil
}

func FromMongoDoc[T any](val primitive.M) (T, error) {
	var result T

	json_data, err := json.Marshal(val)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(json_data, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

const JWT_KEY = "hi mom"

func GenAndSetJWT(c *gin.Context, id string) error {
	jwt_token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": id})

	jwt_token_str, err := jwt_token.SignedString([]byte(JWT_KEY))
	if err != nil {
		return err
	}

	c.Request.Response.Header.Set("token", jwt_token_str)
	return nil
}

func AuthMiddleware(c *gin.Context) {
	jwt_token_str := c.GetHeader("token")

	var user_id string
	ReadBody(c, QPPair{"id", &user_id})
	claims := jwt.MapClaims{"id": user_id}

	token, err := jwt.ParseWithClaims(jwt_token_str, &claims, func(token *jwt.Token) (interface{}, error) {
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

	c.Next()
}
