package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jrallison/go-workers"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	"github.com/topfreegames/extensions/v9/tracing"
	"github.com/topfreegames/khan/mongo"
	"github.com/uber-go/zap"
)

// MongoWorker is the worker that will update mongo
type MongoWorker struct {
	Logger                  zap.Logger
	MongoDB                 interfaces.MongoDB
	MongoCollectionTemplate string
}

// NewMongoWorker creates and returns a new mongo worker
func NewMongoWorker(logger zap.Logger, config *viper.Viper) *MongoWorker {
	w := &MongoWorker{
		Logger: logger,
	}
	w.configureMongoWorker(config)
	return w
}

func (w *MongoWorker) configureMongoWorker(config *viper.Viper) {
	w.MongoCollectionTemplate = config.GetString("mongodb.collectionTemplate")
	w.MongoDB = mongo.GetConfiguredMongoClient()
}

// PerformUpdateMongo updates the clan into mongodb
func (w *MongoWorker) PerformUpdateMongo(m *workers.Msg) {
	tags := opentracing.Tags{"component": "go-workers"}
	span := opentracing.StartSpan("PerformUpdateMongo", tags)
	defer span.Finish()
	defer tracing.LogPanic(span)
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	item := m.Args()
	data := item.MustMap()
	game := data["game"].(string)
	op := data["op"].(string)
	clan := data["clan"].(map[string]interface{})
	clanID := data["clanID"].(string)

	w.updateClanIntoMongoDB(ctx, game, op, clan, clanID)
}

// InsertGame creates a game inside Mongo
func (w *MongoWorker) InsertGame(ctx context.Context, gameID string, clan *Clan) error {
	clanWithNamePrefixes := clan.NewClanWithNamePrefixes()
	clanJSON, err := json.Marshal(clanWithNamePrefixes)
	if err != nil {
		return errors.New("Could not serialize clan")
	}

	var clanMap map[string]interface{}
	json.Unmarshal(clanJSON, &clanMap)

	w.updateClanIntoMongoDB(ctx, gameID, "update", clanMap, clan.PublicID)

	return nil
}

func (w *MongoWorker) updateClanIntoMongoDB(
	ctx context.Context, gameID string, op string, clan map[string]interface{}, clanID string,
) {

	logger := w.Logger.With(
		zap.String("game", gameID),
		zap.String("operation", op),
		zap.String("clanId", clanID),
		zap.String("source", "PerformUpdateMongo"),
	)

	if w.MongoDB != nil {
		mongoCol, mongoSess := w.MongoDB.WithContext(ctx).C(fmt.Sprintf(w.MongoCollectionTemplate, gameID))
		defer mongoSess.Close()

		if op == "update" {
			logger.Debug(fmt.Sprintf("updating clan %s into mongodb", clanID))
			info, err := mongoCol.UpsertId(clanID, clan)
			if err != nil {
				panic(err)
			} else {
				logger.Debug(fmt.Sprintf("ChangeInfo: updated %d, removed %d, matched %d", info.Updated, info.Removed, info.Matched))
			}
		} else if op == "delete" {
			logger.Debug(fmt.Sprintf("deleting clan %s from mongodb", clanID))
			err := mongoCol.RemoveId(clanID)
			if err != nil {
				panic(err)
			}
		}
	}
}
