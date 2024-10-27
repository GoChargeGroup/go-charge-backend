package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// NOTE: validating requests is done by gocharge admins manually (just change it in the DB).
func HandleStationRequest(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	if user_claim.Role != OWNER_ROLE {
		c.JSON(http.StatusUnauthorized, "Only owner accounts are allowed to request charging stations")
		return
	}

	station_data, err := ReadBodyToStruct[NewStationInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user_id, err := primitive.ObjectIDFromHex(user_claim.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	new_station := Station{
		ID:               primitive.NewObjectID(),
		OwnerID:          user_id,
		PictureURLs:      []string{},
		Name:             station_data.Name,
		Description:      station_data.Description,
		Coordinates:      station_data.Coordinates,
		IsPublic:         false,
		OperationalHours: station_data.OperationalHours,
		Address:          station_data.Address,
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

const TRUE_MAX_RESULTS = 20

// Get k-closest stations to a location
func HandleClosestStations(c *gin.Context) {
	body_data, err := ReadBodyToStruct[FindStationsInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	max_results := min(body_data.MaxResults, TRUE_MAX_RESULTS)

	// current_time := time.Now().Unix()
	// current_day_epoch := current_time % (60 * 60 * 24)
	// day_of_week := strconv.Itoa(int(time.Now().Weekday()))
	// start_key := "operational_hours." + day_of_week + ".0"
	// end_key := "operational_hours." + day_of_week + ".1"

	stations, err := Aggregate[FindStationsOutput](STATION_COLL, bson.A{
		// get closest stations from nearest to farthest
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
		// join valid chargers
		bson.D{
			{"$lookup", bson.D{
				{"from", "Chargers"},
				{"localField", "_id"},
				{"foreignField", "station_id"},
				{"as", "chargers"},
				{"pipeline", bson.A{
					bson.D{
						{"$match", bson.D{
							{"status", body_data.Status},
							{"charger_types_id", body_data.PlugType},
							{"kWh_types_id", body_data.PowerOutput},
							{"price", bson.D{
								{"$lte", body_data.MaxPrice},
							}},
						}},
					},
				}},
			}},
		},
		// only return stations with valid chargers.
		bson.D{
			{"$match", bson.D{
				{"chargers.0", bson.D{
					{"$exists", true},
				}},
			}},
		},
		// limit to max results
		bson.D{
			{"$limit", max_results},
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, stations)
}

func HandleFavoriteStation(c *gin.Context) {
	user_claim := c.MustGet(MW_USER_KEY).(UserClaim)
	if user_claim.Role != USER_ROLE {
		c.JSON(http.StatusUnauthorized, "Only user accounts are allowed to favorite a station")
		return
	}

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
	if user_claim.Role != USER_ROLE {
		c.JSON(http.StatusUnauthorized, "Only user accounts are allowed to favorite a station")
		return
	}

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
