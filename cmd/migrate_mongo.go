package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/globalsign/mgo/bson"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	idb "github.com/topfreegames/extensions/gorp/interfaces"
	mongoext "github.com/topfreegames/extensions/mongo"
	imongo "github.com/topfreegames/extensions/mongo/interfaces"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/mongo"
)

var gameID string

var migrateMongoCmd = &cobra.Command{
	Use:   "migrate-mongo",
	Short: "creates MongoDB indexes used by Khan for one game",
	Long: `Creates all indexes used by Khan for one game into a remote MongoDB instance.
If the game do not exists in the main Postgres database, no actions take place.`,
	Run: func(cmd *cobra.Command, args []string) {
		// read config
		config, err := newConfig()
		if err != nil {
			exitWithError(err)
		}

		// connect to main db and check if game exists
		db, err := newDatabase(config)
		if err != nil {
			exitWithError(err)
		}
		game, err := models.GetGameByPublicID(db, gameID)
		if err != nil {
			exitWithError(err)
		}
		if game == nil || game.PublicID != gameID {
			exitWithError(&models.ModelNotFoundError{
				Type: "Game",
				ID:   gameID,
			})
		}

		// connect to mongo and run migrations
		mongoDB, err := newMongo(config)
		if err != nil {
			exitWithError(err)
		}
		err = runMigrations(mongoDB)
		if err != nil {
			exitWithError(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(migrateMongoCmd)

	migrateMongoCmd.Flags().StringVarP(
		&gameID,
		"game",
		"g",
		"",
		"game public ID in main database",
	)
}

func newConfig() (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigType("yaml")
	config.SetConfigFile(ConfigFile)
	config.AddConfigPath(".")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
	return config, config.ReadInConfig()
}

func newDatabase(config *viper.Viper) (idb.Database, error) {
	host := config.GetString("postgres.host")
	user := config.GetString("postgres.user")
	dbName := config.GetString("postgres.dbname")
	password := config.GetString("postgres.password")
	port := config.GetInt("postgres.port")
	sslMode := config.GetString("postgres.sslMode")
	return models.InitDb(host, user, port, sslMode, dbName, password)
}

func newMongo(config *viper.Viper) (imongo.MongoDB, error) {
	config.Set("mongodb.database", config.GetString("mongodb.databaseName"))
	mongoDB, err := mongoext.NewClient("mongodb" /* conf keys prefix */, config)
	if err != nil {
		return nil, err
	}
	return mongoDB.MongoDB, nil
}

func runMigrations(mongoDB imongo.MongoDB) error {
	fmt.Printf(">>>> Running migrations for %s...\n", gameID)
	err := createClanNameTextIndex(mongoDB)
	if err != nil {
		return err
	}
	fmt.Println(">>>> Done.")
	return nil
}

func createClanNameTextIndex(mongoDB imongo.MongoDB) error {
	cmd := mongo.GetClanNameTextIndexCommand(gameID, false)
	var res struct {
		OK               int `bson:"ok"`
		NumIndexesBefore int `bson:"numIndexesBefore"`
		NumIndexesAfter  int `bson:"numIndexesAfter"`
	}
	err := mongoDB.Run(cmd, &res)
	if err != nil {
		return err
	}
	if res.OK != 1 {
		return &MongoCommandError{cmd: cmd}
	}
	if res.NumIndexesAfter == res.NumIndexesBefore {
		fmt.Println(">>>> Clan name text index already exists.")
	}
	return nil
}

// MongoCommandError represents a MongoDB run command error.
type MongoCommandError struct {
	cmd bson.D
}

func (e *MongoCommandError) Error() string {
	return fmt.Sprintf("Error in mongo command: %v", e.cmd)
}

func exitWithError(err error) {
	fmt.Println("panic:", err.Error())
	os.Exit(1)
}
