// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	workers "github.com/jrallison/go-workers"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/queues"
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
)

//Dispatcher is responsible for sending web hooks to workers
type Dispatcher struct {
	app *App
}

//NewDispatcher creates a new dispatcher available to our app
func NewDispatcher(app *App) (*Dispatcher, error) {
	d := &Dispatcher{app: app}
	return d, nil
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

	workers.Enqueue(queues.KhanQueue, "Add", map[string]interface{}{
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

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// PerformDispatchHook dispatches web hooks for a specific game and event type
func (d *Dispatcher) PerformDispatchHook(m *workers.Msg) {
	app := d.app

	item := m.Args()
	data := item.MustMap()

	gameID := data["gameID"].(string)
	eventType, _ := data["eventType"].(json.Number).Int64()
	payload := data["payload"].(map[string]interface{})

	l := d.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "PerformDispatchHook"),
		zap.String("gameID", gameID),
		zap.Int64("eventType", eventType),
	)

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
			Name: fmt.Sprintf("khan-%s", util.VERSION),
		}

		requestURL, err := d.interpolateURL(hook.URL, payload)
		if err != nil {
			app.addError()
			log.E(l, "Could not interpolate webhook.", func(cm log.CM) {
				cm.Write(
					zap.String("requestURL", hook.URL),
					zap.Error(err),
				)
			})
			continue
		}

		payloadJSON, _ := json.Marshal(payload)

		log.D(l, "Requesting Hook URL...", func(cm log.CM) {
			cm.Write(zap.String("requestURL", requestURL))
		})
		req := fasthttp.AcquireRequest()
		req.Header.SetMethod("POST")
		req.AppendBody(payloadJSON)

		parsedURL, err := url.Parse(requestURL)
		if err != nil {
			app.addError()
			log.E(l, "Could not parse request requestURL.", func(cm log.CM) {
				cm.Write(
					zap.String(requestURL, hook.URL),
					zap.Error(err),
				)
			})
		}
		if parsedURL.User != nil {
			username := parsedURL.User.Username()
			password, setten := parsedURL.User.Password()
			if setten == false {
				password = ""
			}
			requestURL = fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.RequestURI())
			req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
		}

		req.SetRequestURI(requestURL)
		resp := fasthttp.AcquireResponse()

		err = client.DoTimeout(req, resp, timeout)
		if err != nil {
			app.addError()
			log.E(l, "Could not request webhook.", func(cm log.CM) {
				cm.Write(zap.String("requestURL", hook.URL), zap.Error(err))
			})
			continue
		}

		if resp.StatusCode() > 399 {
			app.addError()
			log.E(l, "Could not request webhook.", func(cm log.CM) {
				cm.Write(
					zap.String("requestURL", hook.URL),
					zap.Int("statusCode", resp.StatusCode()),
					zap.String("body", string(resp.Body())),
				)
			})
			continue
		}

		log.I(l, "Webhook requested successfully.", func(cm log.CM) {
			cm.Write(
				zap.Int("statusCode", resp.StatusCode()),
				zap.String("requestURL", requestURL),
				zap.String("body", string(resp.Body())),
			)
		})
	}

	return
}
