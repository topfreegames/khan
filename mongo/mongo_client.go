package mongo

import (
	"sync"

	"github.com/spf13/viper"
	"github.com/uber-go/zap"
	mgo "gopkg.in/mgo.v2"
)

var once sync.Once
var mongo *mgo.Database
var mongoSession *mgo.Session

// GetMongo gets a mongo database model
func GetMongo(logger zap.Logger, config *viper.Viper) (*mgo.Database, error) {
	once.Do(func() {
		var err error
		mongoURL := config.GetString("mongodb.url")
		mongoDB := config.GetString("mongodb.databaseName")
		mongoSession, err = mgo.Dial(mongoURL)
		if err != nil {
			logger.Error(err.Error())
		}
		mongo = mongoSession.DB(mongoDB)
		logger.Info("mongo client configured successfully")
	})
	return mongo, nil
}

// GetConfiguredMongoClient gets a configured mongo client
func GetConfiguredMongoClient() (*mgo.Database, *mgo.Session) {
	return mongo, mongoSession
}
