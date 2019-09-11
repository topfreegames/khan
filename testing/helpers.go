package testing

import (
	"fmt"
	"time"

	"github.com/topfreegames/khan/caches"

	"github.com/globalsign/mgo/bson"
	gocache "github.com/patrickmn/go-cache"
	"github.com/topfreegames/extensions/mongo/interfaces"
	"github.com/topfreegames/khan/models"
)

// CreateClanNameTextIndexInMongo creates the necessary text index for clan search in mongo
func CreateClanNameTextIndexInMongo(getTestMongo func() (interfaces.MongoDB, error), gameID string) error {
	mongo, err := getTestMongo()
	if err != nil {
		return err
	}

	cmd := bson.D{
		{Name: "createIndexes", Value: fmt.Sprintf("clans_%s", gameID)},
		{Name: "indexes", Value: []interface{}{
			bson.M{
				"key": bson.M{
					"name":         "text",
					"namePrefixes": "text",
				},
				"name": fmt.Sprintf("clans_%s_name_text_namePrefixes_text_index", gameID),
			},
		}},
	}
	return mongo.Run(cmd, nil)
}

// GetTestDB returns a connection to the test database.
func GetTestDB() (models.DB, error) {
	return models.GetDB(
		"localhost", // host
		"khan_test", // user
		5433,        // port
		"disable",   // sslMode
		"khan_test", // dbName
		"",          // password
	)
}

// GetTestClansSummariesCache returns a test cache for clans summaries.
func GetTestClansSummariesCache() *caches.ClansSummaries {
	return &caches.ClansSummaries{
		Cache:          gocache.New(time.Minute, time.Minute),
		TTL:            time.Minute,
		TTLRandomError: time.Minute / 2,
	}
}
