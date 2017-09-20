package models

import (
	"fmt"

	"github.com/jrallison/go-workers"
	"github.com/spf13/viper"
	"github.com/topfreegames/khan/mongo"
	"github.com/uber-go/zap"
	"gopkg.in/mgo.v2"
)

// MongoWorker is the worker that will update mongo
type MongoWorker struct {
	Logger                  zap.Logger
	MongoDB                 *mgo.Database
	MongoSession            *mgo.Session
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
	w.MongoDB, w.MongoSession = mongo.GetConfiguredMongoClient()
}

// PerformUpdateMongo updates the clan into elasticsearc
func (w *MongoWorker) PerformUpdateMongo(m *workers.Msg) {
	item := m.Args()
	data := item.MustMap()
	mongoSess := w.MongoSession.Clone()
	defer mongoSess.Close()

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
		mongoCol := w.MongoDB.With(mongoSess).C(fmt.Sprintf(w.MongoCollectionTemplate, game))
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
			err := mongoCol.With(mongoSess).RemoveId(clanID)
			if err != nil {
				panic(err)
			}
		}
	}

	return
}
