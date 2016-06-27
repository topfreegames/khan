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
	"time"

	"github.com/golang/glog"
	"github.com/topfreegames/khan/util"
	"github.com/valyala/fasthttp"
)

//Dispatch represents an event hook to be sent to all available dispatchers
type Dispatch struct {
	gameID    string
	eventType int
	payload   []byte
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
	d.workerQueue = make(chan chan Dispatch, d.workerCount)
	d.workQueue = make(chan Dispatch, d.bufferSize)

	// Now, create all of our workers.
	for i := 0; i < d.workerCount; i++ {
		worker := d.newWorker(i+1, d.workerQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case work := <-d.workQueue:
				glog.Info("Received work request")
				go func() {
					worker := <-d.workerQueue

					glog.Info("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()
}

//Wait blocks until all jobs are done
func (d *Dispatcher) Wait() {
	for d.Jobs > 0 {
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
func (d *Dispatcher) DispatchHook(gameID string, eventType int, payload util.JSON) {
	payloadJSON, _ := json.Marshal(payload)
	defer d.startJob()
	work := Dispatch{gameID: gameID, eventType: eventType, payload: payloadJSON}
	// Push the work onto the queue.
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
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
				glog.Info("worker%d: Received work request for game %s event %d with payload %v \n", w.ID, work.gameID, work.eventType, string(work.payload))
				w.handleJob(work)
			}
		}
	}()
}

// DispatchHook dispatches web hooks for a specific game and event type
func (w *Worker) DispatchHook(d Dispatch) error {
	app := w.App
	hooks := app.GetHooks()
	if _, ok := hooks[d.gameID]; !ok {
		return nil
	}
	if _, ok := hooks[d.gameID][d.eventType]; !ok {
		return nil
	}

	timeout := time.Duration(app.Config.GetInt("webhooks.timeout")) * time.Second

	for _, hook := range hooks[d.gameID][d.eventType] {
		glog.Infof("Sending webhook to %s", hook.URL)

		client := fasthttp.Client{
			Name: fmt.Sprintf("khan-%s", VERSION),
		}

		req := fasthttp.AcquireRequest()
		req.Header.SetMethod("POST")
		req.SetRequestURI(hook.URL)
		req.AppendBody(d.payload)
		resp := fasthttp.AcquireResponse()
		err := client.DoTimeout(req, resp, timeout)
		if err != nil {
			w.App.addError()
			glog.Error(fmt.Sprintf("Could not request webhook %s: %s", hook.URL, err.Error()))
			continue
		}
		if resp.StatusCode() > 399 {
			w.App.addError()
			glog.Error(fmt.Sprintf(
				"Could not request webhook %s (status code: %d): %s",
				hook.URL,
				resp.StatusCode(),
				resp.Body(),
			))
			continue
		}
	}

	return nil
}
