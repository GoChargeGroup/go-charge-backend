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
	query_data, err := ReadBodyToStruct[FindStationsInput](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	max_results := min(query_data.MaxResults, TRUE_MAX_RESULTS)

	// current_time := time.Now().Unix()
	// current_day_epoch := current_time % (60 * 60 * 24)
	// day_of_week := strconv.Itoa(int(time.Now().Weekday()))
	// start_key := "operational_hours." + day_of_week + ".0"
	// end_key := "operational_hours." + day_of_week + ".1"

	mongoDBHQ := bson.D{
		{"type", "Point"},
		{"coordinates", query_data.Coordinates},
	}
	stations, err := GetStations(bson.D{
		{"coordinates", bson.D{
			{"$near", bson.D{
				{"$geometry", mongoDBHQ},
				{"$maxDistance", query_data.MaxRadius},
			}},
		}},
		{"is_public", true},
		// {start_key, bson.D{{"$lt", current_day_epoch}}},
		// {end_key, bson.D{{"$gt", current_day_epoch}}},
	}, max_results)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := FindStationsOutput{stations}
	c.JSON(http.StatusOK, result)
}
