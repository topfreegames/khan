// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
	"gopkg.in/gorp.v1"
)

// TransactionMiddleware wraps transactions around the request
type TransactionMiddleware struct {
	App *App
}

// Serve Automatically wrap transaction around the request
func (m *TransactionMiddleware) Serve(c *iris.Context) {
	c.Set("db", m.App.Db)

	tx, err := (m.App.Db).(*gorp.DbMap).Begin()
	if err == nil {
		c.Set("db", tx)
		c.Next()

		if c.Response.StatusCode() > 399 {
			tx.Rollback()
			return
		}

		tx.Commit()
		c.Set("db", m.App.Db)
	} else {
		c.SetStatusCode(500)
		c.Write("Internal server error")
	}
}

// GetCtxDB returns the proper database connection depending on the request context
func GetCtxDB(ctx *iris.Context) (models.DB, error) {
	val := ctx.Get("db")
	if val != nil {
		return val.(models.DB), nil
	}

	return nil, fmt.Errorf("Could not find database instance in request context.")
}

//VersionMiddleware automatically adds a version header to response
type VersionMiddleware struct {
	App *App
}

// Serve automatically adds a version header to response
func (m *VersionMiddleware) Serve(c *iris.Context) {
	c.SetHeader("KHAN-VERSION", VERSION)
	c.Next()
}

//RecoveryMiddleware recovers from errors in Iris
type RecoveryMiddleware struct {
	OnError func(error, []byte)
}

//Serve executes on error handler when errors happen
func (r RecoveryMiddleware) Serve(ctx *iris.Context) {
	defer func() {
		if err := recover(); err != nil {
			if r.OnError != nil {
				r.OnError(err.(error), debug.Stack())
			}
			ctx.Panic()
		}
	}()
	ctx.Next()
}

//LoggerMiddleware is responsible for logging to Zap all requests
type LoggerMiddleware struct {
	Logger zap.Logger
}

// Serve serves the middleware
func (l *LoggerMiddleware) Serve(ctx *iris.Context) {
	log := l.Logger.With(
		zap.String("source", "request"),
	)

	//all except latency to string
	var ip, method, path string
	var status int
	var latency time.Duration
	var startTime, endTime time.Time

	path = ctx.PathString()
	method = ctx.MethodString()

	startTime = time.Now()

	ctx.Next()

	//no time.Since in order to format it well after
	endTime = time.Now()
	latency = endTime.Sub(startTime)

	status = ctx.Response.StatusCode()
	ip = ctx.RemoteAddr()

	reqLog := log.With(
		zap.Time("endTime", endTime),
		zap.Int("statusCode", status),
		zap.Duration("latency", latency),
		zap.String("ip", ip),
		zap.String("method", method),
		zap.String("path", path),
	)

	//finally print the logs
	if status > 399 {
		reqLog.Warn("Request failed.")
		return
	}
	reqLog.Debug("Request successful.")
}

// NewLoggerMiddleware returns the logger middleware
func NewLoggerMiddleware(theLogger zap.Logger) iris.HandlerFunc {
	l := &LoggerMiddleware{Logger: theLogger}
	return l.Serve
}
