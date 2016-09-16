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
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
)

//Dispatch represents an event hook to be sent to all available dispatchers
type Dispatch struct {
	gameID      string
	eventType   int
	payload     map[string]interface{}
	payloadJSON []byte
}

//Dispatcher is responsible for sending web hooks to workers
type Dispatcher struct {
	app         *App
	bufferSize  int
	workerCount int
	Jobs        int
	workQueue   chan Dispatch
	workerQueue chan chan Dispatch
}

//Worker is an unit of work that keeps processing dispatches
type Worker struct {
	ID          int
	App         *App
	Dispatcher  *Dispatcher
	Work        chan Dispatch
	WorkerQueue chan chan Dispatch
}

//NewDispatcher creates a new dispatcher available to our app
func NewDispatcher(app *App, workerCount, bufferSize int) (*Dispatcher, error) {
	return &Dispatcher{app: app, workerCount: workerCount, bufferSize: bufferSize}, nil
}

//Start "starts" the dispatcher
func (d *Dispatcher) Start() {
	l := d.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "Start"),
		zap.Int("workerCount", d.workerCount),
	)

log.D(	l, "Starting dispatcher...")
	d.workerQueue = make(chan chan Dispatch, d.workerCount)
	d.workQueue = make(chan Dispatch, d.bufferSize)

	// Now, create all of our workers.
	for i := 0; i < d.workerCount; i++ {
log.D(		l, "Starting worker...", zap.Int("workerId", i+1))
		worker := d.newWorker(i+1, d.workerQueue)
		worker.Start()
log.D(		l, "Worker started successfully.", zap.Int("workerId", i+1))
	}

	go func() {
		for {
			select {
			case work := <-d.workQueue:
log.I(				l, 
					"Received work request.",
					zap.String("gameID", work.gameID),
					zap.Int("eventType", work.eventType),
				)
				go func() {
log.D(					l, "Waiting for available worker...")
					worker := <-d.workerQueue
log.D(					l, "Worker found! Dispatching work request...")
					worker <- work
				}()
			}
		}
	}()
}

//Wait blocks until all jobs are done
func (d *Dispatcher) Wait(timeout ...int) {
	actualTimeout := 0
	if timeout != nil && len(timeout) == 1 {
		actualTimeout = timeout[0]
	}

	l := d.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "Wait"),
		zap.Int("timeout", actualTimeout),
	)

	start := time.Now()
	timeoutDuration := time.Duration(actualTimeout) * time.Millisecond
	for d.Jobs > 0 {
log.D(		l, "Waiting for jobs to finish...", zap.Int("jobCount", d.Jobs))
		curr := time.Now().Sub(start)
		if actualTimeout > 0 && curr > timeoutDuration {
			log.W(l, "Timeout waiting for jobs to finish.", zap.Duration("timeoutEllapsed", curr))
			break
		}
		time.Sleep(time.Millisecond)
	}
}

func (d *Dispatcher) startJob() {
	d.Jobs++
}

func (d *Dispatcher) finishJob() {
	d.Jobs--
}

//DispatchHook dispatches an event hook for eventType to gameID with the specified payload
func (d *Dispatcher) DispatchHook(gameID string, eventType int, payload map[string]interface{}) {
	payload["type"] = eventType
	payload["id"] = uuid.NewV4()
	payload["timestamp"] = time.Now().Format(time.RFC3339)

	payloadJSON, _ := json.Marshal(payload)
	defer d.startJob()
	work := Dispatch{gameID: gameID, eventType: eventType, payload: payload, payloadJSON: payloadJSON}
	// Push the work onto the queue.
log.D(	d.app.Logger, 
		"Pushing work into dispatch queue.",
		zap.String("source", "dispatcher"),
		zap.String("operation", "DispatchHook"),
	)
	d.workQueue <- work
}

func (d *Dispatcher) newWorker(id int, workerQueue chan chan Dispatch) Worker {
	worker := Worker{
		App:         d.app,
		Dispatcher:  d,
		ID:          id,
		Work:        make(chan Dispatch),
		WorkerQueue: workerQueue,
	}

	return worker
}

func (w *Worker) handleJob(work Dispatch) {
	defer w.Dispatcher.finishJob()
	w.DispatchHook(work)
}

// Start "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) Start() {
	l := w.Dispatcher.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "Start"),
	)

	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
log.I(				l, 
					"Received work request for game.",
					zap.Int("workerID", w.ID),
					zap.String("gameID", work.gameID),
					zap.Int("eventType", work.eventType),
					zap.String("payload", string(work.payloadJSON)),
				)
				w.handleJob(work)
			}
		}
	}()
}

func (w *Worker) interpolateURL(sourceURL string, payload map[string]interface{}) (string, error) {
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

// DispatchHook dispatches web hooks for a specific game and event type
func (w *Worker) DispatchHook(d Dispatch) error {
	l := w.Dispatcher.app.Logger.With(
		zap.String("source", "dispatcher"),
		zap.String("operation", "DispatchHook"),
		zap.String("gameID", d.gameID),
		zap.Int("eventType", d.eventType),
	)

	app := w.App
	hooks := app.GetHooks()
	if _, ok := hooks[d.gameID]; !ok {
log.D(		l, "No hooks found for game.")
		return nil
	}
	if _, ok := hooks[d.gameID][d.eventType]; !ok {
log.D(		l, "No hooks found for event in specified game.")
		return nil
	}

	timeout := time.Duration(app.Config.GetInt("webhooks.timeout")) * time.Second

	for _, hook := range hooks[d.gameID][d.eventType] {
log.I(		w.Dispatcher.app.Logger, "Sending webhook...", zap.String("url", hook.URL))

		client := fasthttp.Client{
			Name: fmt.Sprintf("khan-%s", VERSION),
		}

		url, err := w.interpolateURL(hook.URL, d.payload)
		if err != nil {
			w.App.addError()
			log.E(l, 
				"Could not interpolate webhook.",
				zap.String("url", hook.URL),
				zap.Error(err),
			)
			continue
		}

log.D(		l, "Requesting Hook URL...", zap.String("url", url))
		req := fasthttp.AcquireRequest()
		req.Header.SetMethod("POST")
		req.SetRequestURI(url)
		req.AppendBody(d.payloadJSON)
		resp := fasthttp.AcquireResponse()

		err = client.DoTimeout(req, resp, timeout)
		if err != nil {
			w.App.addError()
			log.E(l, 
				"Could not request webhook.",
				zap.String("url", hook.URL),
				zap.Error(err),
			)
			continue
		}

		if resp.StatusCode() > 399 {
			w.App.addError()
			log.E(l, 
				"Could not request webhook.",
				zap.String("url", hook.URL),
				zap.Int("statusCode", resp.StatusCode()),
				zap.String("body", string(resp.Body())),
			)
			continue
		}

log.I(		l, 
			"Webhook requested successfully.",
			zap.Int("statusCode", resp.StatusCode()),
			zap.String("url", url),
			zap.String("body", string(resp.Body())),
		)
	}

	return nil
}
