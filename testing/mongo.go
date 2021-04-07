package testing

import (
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	"github.com/topfreegames/khan/mongo"
)

// GetTestMongo returns a mongo instance for testing
func GetTestMongo() (interfaces.MongoDB, error) {
	config := viper.New()
	config.SetConfigType("yaml")
	config.SetConfigFile("../config/test.yaml")
	err := config.ReadInConfig()
	if err != nil {
		return nil, err
	}
	logger := NewMockLogger()
	return mongo.GetMongo(logger, config)
}
