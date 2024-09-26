package main

import (
	"encoding/json"
	"errors"

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
