package main

import (
	"log"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// NOTE: validating requests is done by gocharge admins manually (just change it in the DB).
func HandleStationRequest(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	station_data, err := ReadBodyToStruct[NewStationInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	new_station := Station{
		ID:               primitive.NewObjectID(),
		OwnerID:          user_id,
		PictureURLs:      []string{},
		Name:             station_data.Name,
		Description:      station_data.Description,
		Coordinates:      station_data.Coordinates,
		Address:          station_data.Address,
		IsPublic:         false,
		IsDenied:         false,
		OperationalHours: station_data.OperationalHours,
		ReviewCount:      0,
		ReviewScore:      0,
	}
	station_id, err := CreateStation(new_station)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, "A station with this location must already exist")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	station, err := GetStation(bson.D{{"_id", station_id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// note: this should be parallelized, but i cant really bother.
	chargers := []Charger{}
	for _, charger := range station_data.Chargers {
		new_charger := Charger{
			ID:             primitive.NewObjectID(),
			StationID:      station_id,
			Name:           charger.Name,
			Description:    charger.Description,
			KWhTypesId:     charger.KWhTypesId,
			ChargerTypesId: charger.ChargerTypesId,
			Status:         "working",
			Price:          charger.Price,
			TotalPayments:  0,
		}
		charger_id, err := CreateCharger(new_charger)
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, "A station with this location must already exist")
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		charger, err := GetCharger(bson.D{{"_id", charger_id}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		chargers = append(chargers, charger)
	}

	new_station_out := NewStationOutput{station, chargers}
	c.JSON(http.StatusOK, new_station_out)
}

func HandleStationRequestApproval(c *gin.Context) {
	station_data, err := ReadBodyToStruct[ApprovedStationInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	station_id, err := primitive.ObjectIDFromHex(station_data.StationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = UpdateOne(
		STATION_COLL,
		bson.D{
			{"_id", station_id},
			{"is_public", false},
		},
		bson.D{
			{"$set", bson.D{
				{"is_public", station_data.Approved},
				{"is_denied", !station_data.Approved},
			}},
		},
	)
	if err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {
			c.JSON(http.StatusConflict, "A station with this id or non-public status was not found")
		} else {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, "")
}

func HandleUnapprovedStations(c *gin.Context) {
	stations, err := Aggregate[UnapprovedStationsOutput](STATION_COLL, bson.A{
		// only see private (unapproved) stations
		bson.D{
			{"$match", bson.D{
				{"is_public", false},
				{"is_denied", false},
			}},
		},
		// join valid chargers
		bson.D{
			{"$lookup", bson.D{
				{"from", "Chargers"},
				{"localField", "_id"},
				{"foreignField", "station_id"},
				{"as", "chargers"},
			}},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, stations)
}

const TRUE_MAX_RESULTS = 20

// Get k-closest stations to a location. k is capped at 20.
func HandleClosestStations(c *gin.Context) {
	body_data, err := ReadBodyToStruct[FindStationsInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	max_results := min(body_data.MaxResults, TRUE_MAX_RESULTS)

	if body_data.MaxPrice == 0 {
		body_data.MaxPrice = math.MaxFloat64
	}
	if body_data.MaxRadius == 0 {
		body_data.MaxRadius = math.MaxFloat64
	}

	pipeline := bson.A{
		bson.D{
			{"$geoNear", bson.D{
				{"near", bson.D{
					{"type", "Point"},
					{"coordinates", body_data.Coordinates},
				}},
				{"maxDistance", body_data.MaxRadius},
				{"spherical", true},
				{"key", "coordinates"},
				{"distanceField", "distance"},
			}},
		},
		bson.D{
			{"$match", bson.D{
				{"is_public", true},
				{"$expr", bson.D{
					{"$gte", bson.A{
						bson.D{{"$divide", bson.A{"$review_score", "$review_count"}}},
						body_data.MinRating,
					}},
				}},
			}},
		},
	}

	matchConditions := bson.D{}

	if len(body_data.Statuses) > 0 {
		matchConditions = append(matchConditions, bson.E{"status", bson.D{{"$in", body_data.Statuses}}})
	}
	if len(body_data.PowerOutputs) > 0 {
		matchConditions = append(matchConditions, bson.E{"kWh_types_id", bson.D{{"$in", body_data.PowerOutputs}}})
	}
	if len(body_data.PlugTypes) > 0 {
		matchConditions = append(matchConditions, bson.E{"charger_types_id", bson.D{{"$in", body_data.PlugTypes}}})
	}

	matchConditions = append(matchConditions, bson.E{"price", bson.D{{"$lte", body_data.MaxPrice}}})

	if len(matchConditions) > 0 {
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "Chargers"},
				{"localField", "_id"},
				{"foreignField", "station_id"},
				{"as", "chargers"},
				{"pipeline", bson.A{
					bson.D{
						{"$match", matchConditions},
					},
				}},
			}},
		})
	}

	pipeline = append(pipeline, bson.D{
		{"$match", bson.D{
			{"chargers.0", bson.D{
				{"$exists", true},
			}},
		}},
	})

	pipeline = append(pipeline, bson.D{
		{"$limit", max_results},
	})

	stations, err := Aggregate[FindStationsOutput](STATION_COLL, pipeline)
	if err != nil {
		log.Printf("Error with MongoDB aggregation: %v", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, stations)
}

func HandleFavoriteStation(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	body_data, err := ReadBodyToStruct[FavoriteStationInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	station_id, err := primitive.ObjectIDFromHex(body_data.StationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = UpdateUser(
		bson.D{
			{"_id", user_id},
		},
		bson.D{
			{"$addToSet", bson.D{
				{"favorite_station_ids", station_id},
			}},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"_id", user_id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleUnfavoriteStation(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	body_data, err := ReadBodyToStruct[FavoriteStationInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	station_id, err := primitive.ObjectIDFromHex(body_data.StationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = UpdateUser(
		bson.D{
			{"_id", user_id},
		},
		bson.D{
			{"$pull", bson.D{
				{"favorite_station_ids", station_id},
			}},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user, err := GetUser(bson.D{{"_id", user_id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleGetUserChargers(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	filter := bson.D{{"owner_id", user_id}}
	stations, err := GetStations(filter, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chargers"})
		return
	}

	c.JSON(http.StatusOK, stations)
}

func HandleGetStationAndChargers(c *gin.Context) {
	station_data, err := ReadBodyToStruct[GetStationAndChargersInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	station, err := GetStation(bson.D{{"_id", station_data.StationID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	chargers, err := GetAll[Charger](CHARGER_COLL, bson.D{{"station_id", station_data.StationID}}, TRUE_MAX_RESULTS)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	station_and_chargers := GetStationAndChargersOutput{station, chargers}
	c.JSON(http.StatusOK, station_and_chargers)
}

func HandleEditStation(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	body_data, err := ReadBodyToStruct[EditStationInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// edit station
	err = UpdateOne(
		STATION_COLL,
		bson.D{
			{"_id", body_data.ID},
			{"owner_id", user_id}, // ensure owner owns this station
		},
		bson.D{
			{"picture_urls", body_data.PictureURLs},
			{"name", body_data.Name},
			{"description", body_data.Description},
			{"coordinates", body_data.Coordinates},
			{"address", body_data.Address},
			{"operational_hours", body_data.OperationalHours},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// get and return updated station doc
	station, err := GetStation(bson.D{{"_id", body_data.ID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, station)
}
