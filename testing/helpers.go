package testing

import (
	"time"

	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/caches"
	"github.com/topfreegames/khan/util"

	gocache "github.com/patrickmn/go-cache"
	"github.com/topfreegames/khan/models"
)

// // CreateClanNameTextIndexInMongo creates the necessary text index for clan search in mongo
// func CreateClanNameTextIndexInMongo(getTestMongo func() (interfaces.MongoDB, error), gameID string) error {
// 	db, err := getTestMongo()
// 	if err != nil {
// 		return err
// 	}
// 	return db.Run(mongo.GetClanNameTextIndexCommand(gameID, false), nil)
// }

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

// DecryptTestPlayer replaces the encrypted name by the plain text name in the player object
func DecryptTestPlayer(encryptionKey []byte, player *models.Player) {
	name, err := util.DecryptData(player.Name, encryptionKey)
	Expect(err).NotTo(HaveOccurred())
	player.Name = name
}

//UpdateEncryptingTestPlayer encrypt player name and save it to database
func UpdateEncryptingTestPlayer(db models.DB, encryptionKey []byte, player *models.Player) {
	name, err := util.EncryptData(player.Name, encryptionKey)
	Expect(err).NotTo(HaveOccurred())
	player.Name = name
	rows, err := db.Update(player)
	Expect(err).NotTo(HaveOccurred())
	Expect(rows).To(BeEquivalentTo(1))

}
