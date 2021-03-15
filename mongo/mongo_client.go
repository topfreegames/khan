package mongo

import (
	"sync"

	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/mongo"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	"github.com/uber-go/zap"
)

var once sync.Once
var client *mongo.Client

// GetMongo gets a mongo database model
func GetMongo(logger zap.Logger, config *viper.Viper) (interfaces.MongoDB, error) {
	var err error

	once.Do(func() {
		client, err = mongo.NewClient("mongodb", config)
		if err != nil {
			message := err.Error()
			logger.Error(message)
			return
		}

		logger.Info("mongo client configured successfully")
	})

	return client.MongoDB, err
}

// GetConfiguredMongoClient gets a configured mongo client
func GetConfiguredMongoClient() interfaces.MongoDB {
	if client != nil {
		return client.MongoDB
	}
	return nil
}
