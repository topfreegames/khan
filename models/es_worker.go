package models

import (
	"context"
	"time"

	"github.com/jrallison/go-workers"
	"github.com/topfreegames/khan/es"
	"github.com/uber-go/zap"
)

// ESWorker is the worker that will update elasticsearch
type ESWorker struct {
	Logger zap.Logger
	ES     *es.Client
}

// NewESWorker creates and returns a new elasticsearch worker
func NewESWorker(logger zap.Logger) *ESWorker {
	w := &ESWorker{
		Logger: logger,
	}
	w.configureESWorker()
	return w
}

func (w *ESWorker) configureESWorker() {
	w.ES = es.GetConfiguredClient()
}

// PerformUpdateES updates the clan into elasticsearc
func (w *ESWorker) PerformUpdateES(m *workers.Msg) {
	item := m.Args()
	data := item.MustMap()

	index := data["index"].(string)
	op := data["op"].(string)
	clan := data["clan"].(string)
	clanID := data["clanID"].(string)

	l := w.Logger.With(
		zap.String("index", index),
		zap.String("operation", op),
		zap.String("clan", clan),
		zap.String("source", "PerformUpdateES"),
	)

	if w.ES != nil {
		start := time.Now()
		if op == "index" {
			_, err := w.ES.Client.
				Index().
				Index(index).
				Type("clan").
				Id(clanID).
				BodyString(clan).
				Do(context.TODO())
			if err != nil {
				l.Error("Failed to index clan into Elastic Search")
				return
			}
			l.Info("Successfully indexed clan into Elastic Search.", zap.Duration("latency", time.Now().Sub(start)))
		} else if op == "delete" {
			_, err := w.ES.Client.
				Delete().
				Index(index).
				Type("clan").
				Id(clanID).
				Do(context.TODO())
			if err != nil {
				l.Error("Failed to delete clan from Elastic Search.", zap.Error(err))
			}
			l.Info("Successfully deleted clan from Elastic Search.", zap.Duration("latency", time.Now().Sub(start)))
		}
	}

	return
}
