package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
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
