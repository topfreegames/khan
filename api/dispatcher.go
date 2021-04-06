// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	workers "github.com/jrallison/go-workers"
	opentracing "github.com/opentracing/opentracing-go"
	uuid "github.com/satori/go.uuid"
	ehttp "github.com/topfreegames/extensions/v9/http"
	"github.com/topfreegames/extensions/v9/tracing"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/queues"
	"github.com/uber-go/zap"
	"github.com/valyala/fasttemplate"
)

var once sync.Once
var httpClient *http.Client

const hookInternalFailures = "hook_internal_failures"
const requestingHookMilliseconds = "requesting_hook_milliseconds"

//Dispatcher is responsible for sending web hooks to workers
type Dispatcher struct {
	app        *App
	httpClient *http.Client
}

//NewDispatcher creates a new dispatcher available to our app
func NewDispatcher(app *App) (*Dispatcher, error) {
	d := &Dispatcher{app: app}
	d.configureHTTPClient()
	d.httpClient = httpClient
	return d, nil
}

func getHTTPTransport(
	maxIdleConns, maxIdleConnsPerHost int,
) http.RoundTripper {
	if _, ok := http.DefaultTransport.(*http.Transport); !ok {
		return http.DefaultTransport // tests use a mock transport
	}

	// We can't get http.DefaultTransport here and update its
	// fields since it's an exported variable, so other libs could
	// also change it and overwrite. This hardcoded values are copied
	// from http.DefaultTransport but could be configurable too.
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
	}
}

func (d *Dispatcher) configureHTTPClient() {
	timeout := time.Duration(d.app.Config.GetInt("webhooks.timeout")) * time.Millisecond
	maxIdleConns := d.app.Config.GetInt("webhooks.maxIdleConns")
	maxIdleConnsPerHost := d.app.Config.GetInt("webhooks.maxIdleConnsPerHost")

	once.Do(func() {
		httpClient = &http.Client{
			Transport: getHTTPTransport(maxIdleConns, maxIdleConnsPerHost),
			Timeout:   timeout,
		}
		ehttp.Instrument(httpClient)
	})
}

//DispatchHook dispatches an event hook for eventType to gameID with the specified payload
func (d *Dispatcher) DispatchHook(gameID string, eventType int, payload map[string]interface{}) {
	payload["type"] = eventType
	payload["id"] = uuid.NewV4()
	payload["timestamp"] = time.Now().Format(time.RFC3339)

	// Push the work onto the queue.
	log.D(d.app.Logger, "Pushing work into dispatch queue.", func(cm log.CM) {
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
	jtags := opentracing.Tags{"component": "go-workers"}
	span := opentracing.StartSpan("PerformDispatchHook", jtags)
	defer span.Finish()
	defer tracing.LogPanic(span)
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	app := d.app
	statsd := app.DDStatsD

	item := m.Args()
	data := item.MustMap()

	gameID := data["gameID"].(string)
	eventType, _ := data["eventType"].(json.Number).Int64()
	payload := data["payload"].(map[string]interface{})

	logger := d.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "PerformDispatchHook"),
		zap.String("gameID", gameID),
		zap.Int64("eventType", eventType),
	)

	hooks := app.GetHooks(ctx)
	if _, ok := hooks[gameID]; !ok {
		log.D(logger, "No hooks found for game.")
		return
	}
	if _, ok := hooks[gameID][int(eventType)]; !ok {
		log.D(logger, "No hooks found for event in specified game.")
		return
	}

	for _, hook := range hooks[gameID][int(eventType)] {
		log.D(app.Logger, "Sending webhook...", func(cm log.CM) {
			cm.Write(zap.String("url", hook.URL))
		})

		requestURL, err := d.interpolateURL(hook.URL, payload)
		if err != nil {
			app.addError()
			tags := []string{
				"error:true",
				fmt.Sprintf("url:%s", hook.URL),
				fmt.Sprintf("game:%s", gameID),
			}
			statsd.Increment(hookInternalFailures, tags...)

			log.E(logger, "Could not interpolate webhook.", func(cm log.CM) {
				cm.Write(
					zap.String("requestURL", hook.URL),
					zap.Error(err),
				)
			})
			continue
		}

		payloadJSON, _ := json.Marshal(payload)

		log.D(logger, "Requesting Hook URL...", func(cm log.CM) {
			cm.Write(zap.String("requestURL", requestURL))
		})

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.E(logger, "failed to create webhook request", func(cm log.CM) {
				cm.Write(
					zap.String("requestURL", hook.URL),
					zap.Error(err),
				)
			})
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)

		parsedURL, err := url.Parse(requestURL)
		if err != nil {
			app.addError()
			tags := []string{
				"error:true",
				fmt.Sprintf("url:%s", hook.URL),
				fmt.Sprintf("game:%s", gameID),
			}
			statsd.Increment(hookInternalFailures, tags...)

			log.E(logger, "Could not parse request requestURL.", func(cm log.CM) {
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
			req.SetBasicAuth(username, password)
		}

		start := time.Now()
		resp, err := d.httpClient.Do(req)
		if err != nil {
			app.addError()
			tags := []string{
				"error:true",
				fmt.Sprintf("url:%s", hook.URL),
				fmt.Sprintf("game:%s", gameID),
				fmt.Sprintf("status:500"),
			}
			elapsed := time.Since(start)
			statsd.Timing(requestingHookMilliseconds, elapsed, tags...)
			statsd.Increment(hookInternalFailures, tags...)

			log.E(logger, "Could not request webhook.", func(cm log.CM) {
				cm.Write(zap.String("requestURL", hook.URL), zap.Error(err))
			})
			continue
		}
		defer resp.Body.Close()

		body, respErr := ioutil.ReadAll(resp.Body)
		if respErr != nil {
			log.E(logger, "failed to read webhook response", func(cm log.CM) {
				cm.Write(zap.String("requestURL", hook.URL), zap.Error(respErr))
			})
			continue
		}

		tags := []string{
			fmt.Sprintf("error:%t", resp.StatusCode > 399),
			fmt.Sprintf("url:%s", hook.URL),
			fmt.Sprintf("game:%s", gameID),
			fmt.Sprintf("status:%d", resp.StatusCode),
		}
		elapsed := time.Since(start)
		statsd.Timing(requestingHookMilliseconds, elapsed, tags...)

		if resp.StatusCode > 399 {
			app.addError()
			log.E(logger, "Could not request webhook.", func(cm log.CM) {
				cm.Write(
					zap.String("requestURL", hook.URL),
					zap.Int("statusCode", resp.StatusCode),
					zap.String("body", string(body)),
				)
			})
			continue
		}

		log.D(logger, "Webhook requested successfully.", func(cm log.CM) {
			cm.Write(
				zap.Int("statusCode", resp.StatusCode),
				zap.String("requestURL", requestURL),
				zap.String("body", string(body)),
			)
		})
	}

	return
}
