// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	workers "github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/log"
	"github.com/uber-go/zap"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
)

const khanQueue = "khan_webhooks"

//Dispatcher is responsible for sending web hooks to workers
type Dispatcher struct {
	app         *App
	workerCount int
}

//NewDispatcher creates a new dispatcher available to our app
func NewDispatcher(app *App, workerCount int) (*Dispatcher, error) {
	d := &Dispatcher{app: app, workerCount: workerCount}
	d.Configure()
	return d, nil
}

//Configure dispatcher
func (d *Dispatcher) Configure() {
	redisHost := d.app.Config.GetString("redis.host")
	redisPort := d.app.Config.GetInt("redis.port")
	redisDatabase := d.app.Config.GetInt("redis.database")
	redisPool := d.app.Config.GetInt("redis.pool")
	if redisPool == 0 {
		redisPool = 30
	}

	l := d.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "Configure"),
		zap.Int("workerCount", d.workerCount),
		zap.String("redisHost", redisHost),
		zap.Int("redisPort", redisPort),
		zap.Int("redisDatabase", redisDatabase),
		zap.Int("redisPool", redisPool),
	)

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
	redisPass := d.app.Config.GetString("redis.password")
	if redisPass != "" {
		opts["password"] = redisPass
	}
	l.Debug("Configuring worker...")
	workers.Configure(opts)

	workers.Process(khanQueue, d.PerformDispatchHook, d.workerCount)
	l.Info("Worker configured.")
}

//Start "starts" the dispatcher
func (d *Dispatcher) Start() {
	l := d.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "Start"),
	)

	log.D(l, "Starting dispatcher...")
	if d.app.Config.GetBool("webhooks.runStats") {
		jobsStatsPort := d.app.Config.GetInt("webhooks.statsPort")
		go workers.StatsServer(jobsStatsPort)
	}

	workers.Run()
}

//NonblockingStart non-blocking
func (d *Dispatcher) NonblockingStart() {
	workers.Start()
}

//DispatchHook dispatches an event hook for eventType to gameID with the specified payload
func (d *Dispatcher) DispatchHook(gameID string, eventType int, payload map[string]interface{}) {
	payload["type"] = eventType
	payload["id"] = uuid.NewV4()
	payload["timestamp"] = time.Now().Format(time.RFC3339)

	// Push the work onto the queue.
	log.I(d.app.Logger, "Pushing work into dispatch queue.", func(cm log.CM) {
		cm.Write(
			zap.String("source", "dispatcher"),
			zap.String("operation", "DispatchHook"),
		)
	})

	workers.Enqueue(khanQueue, "Add", map[string]interface{}{
		"gameID":    gameID,
		"eventType": eventType,
		"payload":   payload,
	})
}

func (d *Dispatcher) interpolateURL(sourceURL string, payload map[string]interface{}) (string, error) {
	t, err := fasttemplate.NewTemplate(sourceURL, "{{", "}}")
	if err != nil {
		return sourceURL, err
	}
	s := t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		pieces := strings.Split(tag, ".")
		var item interface{}
		item = payload

		if len(pieces) > 1 {
			for _, piece := range pieces {
				switch item.(type) {
				case map[string]interface{}:
					item = item.(map[string]interface{})[piece]
				default:
					return 0, nil
				}
			}
			valEncoded := url.QueryEscape(
				fmt.Sprintf("%v", item),
			)
			return w.Write([]byte(valEncoded))
		}

		if val, ok := payload[tag]; ok {
			valEncoded := url.QueryEscape(
				fmt.Sprintf("%v", val),
			)

			return w.Write([]byte(valEncoded))
		}

		return 0, nil
	})
	return s, nil
}

// PerformDispatchHook dispatches web hooks for a specific game and event type
func (d *Dispatcher) PerformDispatchHook(m *workers.Msg) {
	l := d.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "PerformDispatchHook"),
	)
	app := d.app

	item := m.Args()
	data := item.MustMap()

	gameID := data["gameID"].(string)
	eventType, _ := data["eventType"].(json.Number).Int64()
	payload := data["payload"].(map[string]interface{})

	hooks := app.GetHooks()
	if _, ok := hooks[gameID]; !ok {
		log.D(l, "No hooks found for game.")
		return
	}
	if _, ok := hooks[gameID][int(eventType)]; !ok {
		log.D(l, "No hooks found for event in specified game.")
		return
	}

	timeout := time.Duration(app.Config.GetInt("webhooks.timeout")) * time.Second

	for _, hook := range hooks[gameID][int(eventType)] {
		log.I(app.Logger, "Sending webhook...", func(cm log.CM) {
			cm.Write(zap.String("url", hook.URL))
		})

		client := fasthttp.Client{
			Name: fmt.Sprintf("khan-%s", VERSION),
		}

		url, err := d.interpolateURL(hook.URL, payload)
		if err != nil {
			app.addError()
			log.E(l, "Could not interpolate webhook.", func(cm log.CM) {
				cm.Write(
					zap.String("url", hook.URL),
					zap.Error(err),
				)
			})
			continue
		}

		payloadJSON, _ := json.Marshal(payload)

		log.D(l, "Requesting Hook URL...", func(cm log.CM) {
			cm.Write(zap.String("url", url))
		})
		req := fasthttp.AcquireRequest()
		req.Header.SetMethod("POST")
		req.SetRequestURI(url)
		req.AppendBody(payloadJSON)
		resp := fasthttp.AcquireResponse()

		err = client.DoTimeout(req, resp, timeout)
		if err != nil {
			app.addError()
			log.E(l, "Could not request webhook.", func(cm log.CM) {
				cm.Write(zap.String("url", hook.URL), zap.Error(err))
			})
			continue
		}

		if resp.StatusCode() > 399 {
			app.addError()
			log.E(l, "Could not request webhook.", func(cm log.CM) {
				cm.Write(
					zap.String("url", hook.URL),
					zap.Int("statusCode", resp.StatusCode()),
					zap.String("body", string(resp.Body())),
				)
			})
			continue
		}

		log.I(l, "Webhook requested successfully.", func(cm log.CM) {
			cm.Write(
				zap.Int("statusCode", resp.StatusCode()),
				zap.String("url", url),
				zap.String("body", string(resp.Body())),
			)
		})
	}

	return
}
