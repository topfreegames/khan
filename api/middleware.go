// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"time"

	"github.com/labstack/echo"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/util"
	"github.com/uber-go/zap"
)

func getBodyFromNext(c echo.Context, next echo.HandlerFunc) (string, error) {
	res := c.Response()
	rw := res.Writer()
	buf := new(bytes.Buffer)
	mw := io.MultiWriter(rw, buf)
	res.SetWriter(mw)

	err := next(c)

	body := buf.String()
	return body, err
}

//NewBodyExtractionMiddleware with API version
func NewBodyExtractionMiddleware() *BodyExtractionMiddleware {
	return &BodyExtractionMiddleware{}
}

//BodyExtractionMiddleware extracts the body
type BodyExtractionMiddleware struct{}

// Serve serves the middleware
func (v *BodyExtractionMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := getBodyFromNext(c, next)
		c.Set("body", body)
		return err
	}
}

//NewVersionMiddleware with API version
func NewVersionMiddleware() *VersionMiddleware {
	return &VersionMiddleware{
		Version: util.VERSION,
	}
}

//VersionMiddleware inserts the current version in all requests
type VersionMiddleware struct {
	Version string
}

// Serve serves the middleware
func (v *VersionMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, fmt.Sprintf("Khan/v%s", v.Version))
		c.Response().Header().Set("Khan-Server", fmt.Sprintf("Khan/v%s", v.Version))
		return next(c)
	}
}

func getHTTPParams(ctx echo.Context) (string, map[string]string, string) {
	qs := ""
	if len(ctx.QueryParams()) > 0 {
		qsBytes, _ := json.Marshal(ctx.QueryParams())
		qs = string(qsBytes)
	}

	headers := map[string]string{}
	for _, headerKey := range ctx.Response().Header().Keys() {
		headers[string(headerKey)] = string(ctx.Response().Header().Get(headerKey))
	}

	cookies := string(ctx.Response().Header().Get("Cookie"))
	return qs, headers, cookies
}

//NewRecoveryMiddleware returns a configured middleware
func NewRecoveryMiddleware(onError func(error, []byte)) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		OnError: onError,
	}
}

//RecoveryMiddleware recovers from errors
type RecoveryMiddleware struct {
	OnError func(error, []byte)
}

//Serve executes on error handler when errors happen
func (r *RecoveryMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				eError, ok := err.(error)
				if !ok {
					eError = fmt.Errorf(fmt.Sprintf("%v", err))
				}
				if r.OnError != nil {
					r.OnError(eError, debug.Stack())
				}
				c.Error(eError)
			}
		}()
		return next(c)
	}
}

// NewLoggerMiddleware returns the logger middleware
func NewLoggerMiddleware(theLogger zap.Logger) *LoggerMiddleware {
	l := &LoggerMiddleware{Logger: theLogger}
	return l
}

//LoggerMiddleware is responsible for logging to Zap all requests
type LoggerMiddleware struct {
	Logger zap.Logger
}

// Serve serves the middleware
func (l *LoggerMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := l.Logger.With(
			zap.String("source", "request"),
		)

		//all except latency to string
		var ip, method, path string
		var status int
		var latency time.Duration
		var startTime, endTime time.Time

		path = c.Path()
		method = c.Request().Method()

		startTime = time.Now()

		err := next(c)

		//no time.Since in order to format it well after
		endTime = time.Now()
		latency = endTime.Sub(startTime)

		status = c.Response().Status()
		ip = c.Request().RemoteAddress()

		route := c.Get("route")
		if route == nil {
			log.D(logger, "Route does not have route set in ctx")
			return err
		}

		reqLog := logger.With(
			zap.String("route", route.(string)),
			zap.Time("endTime", endTime),
			zap.Int("statusCode", status),
			zap.Duration("latency", latency),
			zap.String("ip", ip),
			zap.String("method", method),
			zap.String("path", path),
		)

		//request failed
		if status > 399 && status < 500 {
			log.D(reqLog, "Request failed.")
			return err
		}

		//request is ok, but server failed
		if status > 499 {
			log.D(reqLog, "Response failed.")
			return err
		}

		//Everything went ok
		if cm := reqLog.Check(zap.DebugLevel, "Request successful."); cm.OK() {
			cm.Write()
		}

		return err
	}
}
