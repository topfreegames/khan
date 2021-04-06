package cmd

import (
	"fmt"
	"strings"

	"github.com/uber-go/zap"

	"github.com/globalsign/mgo/bson"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	idb "github.com/topfreegames/extensions/v9/gorp/interfaces"
	mongoext "github.com/topfreegames/extensions/v9/mongo"
	imongo "github.com/topfreegames/extensions/v9/mongo/interfaces"
	"github.com/topfreegames/khan/log"
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
		// create logger
		logger := zap.New(zap.NewJSONEncoder(), zap.InfoLevel)
		logger = logger.With(
			zap.String("source", "cmd/migrate_mongo.go"),
			zap.String("operation", "migrateMongoCmd.Run"),
			zap.String("game", gameID),
		)

		// read config
		config, err := newConfig()
		if err != nil {
			log.F(logger, "Error reading config.", func(cm log.CM) {
				cm.Write(zap.String("error", err.Error()))
			})
		}

		// connect to main db and check if game exists
		db, err := newDatabase(config)
		if err != nil {
			log.F(logger, "Error connecting to postgres.", func(cm log.CM) {
				cm.Write(zap.String("error", err.Error()))
			})
		}
		_, err = models.GetGameByPublicID(db, gameID)
		if err != nil {
			log.F(logger, "Error fetching game from postgres.", func(cm log.CM) {
				cm.Write(zap.String("error", err.Error()))
			})
		}

		// connect to mongo and run migrations
		mongoDB, err := newMongo(config)
		if err != nil {
			log.F(logger, "Error connecting to mongo.", func(cm log.CM) {
				cm.Write(zap.String("error", err.Error()))
			})
		}
		err = runMigrations(mongoDB, logger)
		if err != nil {
			log.F(logger, "Error running mongo migrations.", func(cm log.CM) {
				cm.Write(zap.String("error", err.Error()))
			})
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
	config.SetEnvPrefix("khan")
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

func runMigrations(mongoDB imongo.MongoDB, logger zap.Logger) error {
	logger = logger.With(
		zap.String("source", "cmd/migrate_mongo.go"),
		zap.String("operation", "runMigrations"),
		zap.String("game", gameID),
	)

	log.I(logger, "Running mongo migrations for game...")

	// migrations
	type Migration func(imongo.MongoDB, zap.Logger) error
	migrations := []Migration{
		createClanNameTextIndex,
	}
	for _, migration := range migrations {
		if err := migration(mongoDB, logger); err != nil {
			return err
		}
	}

	log.I(logger, "Migrated.")

	return nil
}

func createClanNameTextIndex(mongoDB imongo.MongoDB, logger zap.Logger) error {
	logger = logger.With(
		zap.String("source", "cmd/migrate_mongo.go"),
		zap.String("operation", "createClanNameTextIndex"),
		zap.String("game", gameID),
	)

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
		log.W(logger, "Clan name text index already exists for this game.")
	}
	return nil
}

// MongoCommandError represents a MongoDB run command error.
type MongoCommandError struct {
	cmd bson.D
}

func (e *MongoCommandError) Error() string {
	return fmt.Sprintf("Error in mongo command: %v.", e.cmd)
}
