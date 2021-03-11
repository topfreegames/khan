package testing

import (
	"time"

	"github.com/topfreegames/khan/caches"

	gocache "github.com/patrickmn/go-cache"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/mongo"
)

// CreateClanNameTextIndexInMongo creates the necessary text index for clan search in mongo
func CreateClanNameTextIndexInMongo(getTestMongo func() (interfaces.MongoDB, error), gameID string) error {
	db, err := getTestMongo()
	if err != nil {
		return err
	}
	return db.Run(mongo.GetClanNameTextIndexCommand(gameID, false), nil)
}

var testDB models.DB

// GetTestDB returns a connection to the test database.
func GetTestDB() (models.DB, error) {
	if testDB != nil {
		return testDB, nil
	}
	db, err := models.GetDB(
		"localhost", // host
		"khan_test", // user
		5433,        // port
		"disable",   // sslMode
		"khan_test", // dbName
		"",          // password
	)

	if err != nil {
		return nil, err
	}

	testDB = db

	return testDB, nil
}

// GetTestClansSummariesCache returns a test cache for clans summaries.
func GetTestClansSummariesCache(ttl, cleanupInterval time.Duration) *caches.ClansSummaries {
	return &caches.ClansSummaries{
		Cache: gocache.New(ttl, cleanupInterval),
	}
}
