// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models_test

import (
	"bytes"
	"fmt"
	dlog "log"
	"strconv"

	workers "github.com/jrallison/go-workers"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/mongo/interfaces"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/mongo"
	"github.com/topfreegames/khan/queues"
	kt "github.com/topfreegames/khan/testing"
)

// GetTestDB returns a connection to the test database
func GetTestDB() (models.DB, error) {
	return models.GetDB("localhost", "khan_test", 5433, "disable", "khan_test", "")
}

func GetTestMongo() (interfaces.MongoDB, error) {
	config := viper.New()
	config.SetConfigType("yaml")
	config.SetConfigFile("../config/test.yaml")
	err := config.ReadInConfig()
	if err != nil {
		return nil, err
	}
	l := kt.NewMockLogger()
	return mongo.GetMongo(l, config)
}

// GetFaultyTestDB returns an ill-configured test database
func GetFaultyTestDB() models.DB {
	faultyDb, _ := models.InitDb("localhost", "khan_tet", 5433, "disable", "khan_test", "")
	return faultyDb
}

func ConfigureAndStartGoWorkers() error {
	config := viper.New()
	config.SetConfigType("yaml")
	config.SetConfigFile("../config/test.yaml")
	err := config.ReadInConfig()
	if err != nil {
		return err
	}

	redisHost := config.GetString("redis.host")
	redisPort := config.GetInt("redis.port")
	redisDatabase := config.GetInt("redis.database")
	redisPool := config.GetInt("redis.pool")
	workerCount := config.GetInt("webhooks.workers")
	if redisPool == 0 {
		redisPool = 30
	}

	if workerCount == 0 {
		workerCount = 5
	}

	opts := map[string]string{
		// location of redis instance
		"server": fmt.Sprintf("%s:%d", redisHost, redisPort),
		// instance of the database
		"database": strconv.Itoa(redisDatabase),
		// number of connections to keep open with redis
		"pool": strconv.Itoa(redisPool),
		// unique process id
		"process": uuid.NewV4().String(),
	}
	redisPass := config.GetString("redis.password")
	if redisPass != "" {
		opts["password"] = redisPass
	}
	workers.Configure(opts)
	if config.GetBool("webhooks.logToBuf") {
		var buf bytes.Buffer
		workers.Logger = dlog.New(&buf, "test: ", 0)
	}

	l := kt.NewMockLogger()
	mongoWorker := models.NewMongoWorker(l, config)
	workers.Process(queues.KhanMongoQueue, mongoWorker.PerformUpdateMongo, workerCount)
	workers.Start()
	return nil
}
