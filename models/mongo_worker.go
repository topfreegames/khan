package models

import (
	"context"
	"fmt"

	"github.com/jrallison/go-workers"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/mongo/interfaces"
	"github.com/topfreegames/extensions/tracing"
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

// PerformUpdateMongo updates the clan into elasticsearc
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

	l := w.Logger.With(
		zap.String("game", game),
		zap.String("operation", op),
		zap.String("clanId", clanID),
		zap.String("source", "PerformUpdateMongo"),
	)

	if w.MongoDB != nil {
		mongoCol, mongoSess := w.MongoDB.WithContext(ctx).C(fmt.Sprintf(w.MongoCollectionTemplate, game))
		defer mongoSess.Close()

		if op == "update" {
			l.Debug(fmt.Sprintf("updating clan %s into mongodb", clanID))
			info, err := mongoCol.UpsertId(clanID, clan)
			if err != nil {
				panic(err)
			} else {
				l.Debug(fmt.Sprintf("ChangeInfo: updated %d, removed %d, matched %d", info.Updated, info.Removed, info.Matched))
			}
		} else if op == "delete" {
			l.Debug(fmt.Sprintf("deleting clan %s from mongodb", clanID))
			err := mongoCol.RemoveId(clanID)
			if err != nil {
				panic(err)
			}
		}
	}

	return
}
