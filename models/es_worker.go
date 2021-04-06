package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jrallison/go-workers"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/topfreegames/extensions/v9/tracing"
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
	tags := opentracing.Tags{"component": "go-workers"}
	span := opentracing.StartSpan("PerformUpdateES", tags)
	defer span.Finish()
	defer tracing.LogPanic(span)
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	item := m.Args()
	data := item.MustMap()

	index := data["index"].(string)
	op := data["op"].(string)
	clan := data["clan"].(map[string]interface{})
	clanID := data["clanID"].(string)

	logger := w.Logger.With(
		zap.String("index", index),
		zap.String("operation", op),
		zap.String("clanId", clanID),
		zap.String("source", "PerformUpdateES"),
	)

	if w.ES != nil {
		start := time.Now()
		if op == "index" {
			body, er := json.Marshal(clan)
			if er != nil {
				logger.Error("Failed to get clan JSON and index into Elastic Search", zap.Error(er))
				return
			}
			_, err := w.ES.Client.
				Index().
				Index(index).
				Type("clan").
				Id(clanID).
				BodyString(string(body)).
				Do(ctx)
			if err != nil {
				logger.Error("Failed to index clan into Elastic Search")
				return
			}

			logger.Debug("Successfully indexed clan into Elastic Search.", zap.Duration("latency", time.Now().Sub(start)))
		} else if op == "update" {
			_, err := w.ES.Client.
				Update().
				Index(index).
				Type("clan").
				Id(clanID).
				Doc(clan).
				Do(ctx)
			if err != nil {
				logger.Error("Failed to update clan from Elastic Search.", zap.Error(err))
			}

			logger.Debug("Successfully updated clan from Elastic Search.", zap.Duration("latency", time.Now().Sub(start)))
		} else if op == "delete" {
			_, err := w.ES.Client.
				Delete().
				Index(index).
				Type("clan").
				Id(clanID).
				Do(ctx)

			if err != nil {
				logger.Error("Failed to delete clan from Elastic Search.", zap.Error(err))
			}

			logger.Debug("Successfully deleted clan from Elastic Search.", zap.Duration("latency", time.Now().Sub(start)))
		}
	}

	return
}
