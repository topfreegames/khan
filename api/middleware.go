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

	"github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
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
		return WithSegment("middleware-version", c, func() error {
			c.Response().Header().Set(echo.HeaderServer, fmt.Sprintf("Khan/v%s", v.Version))
			c.Response().Header().Set("Khan-Server", fmt.Sprintf("Khan/v%s", v.Version))
			return next(c)
		})
	}
}

//NewSentryMiddleware returns a new sentry middleware
func NewSentryMiddleware(app *App) *SentryMiddleware {
	return &SentryMiddleware{
		App: app,
	}
}

//SentryMiddleware is responsible for sending all exceptions to sentry
type SentryMiddleware struct {
	App *App
}

// Serve serves the middleware
func (s *SentryMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return WithSegment("middleware-sentry", c, func() error {
			err := next(c)
			body := c.Get("body").(string)

			if err == nil {
				status := c.Response().Status()

				//request is ok, but server failed
				if status > 499 {
					err = fmt.Errorf("Server failed to process response with status code 500: %s", body)
				}
			}

			if err != nil {
				if httpErr, ok := err.(*echo.HTTPError); ok {
					if httpErr.Code < 500 {
						return err
					}
				}
				tags := map[string]string{
					"source": "app",
					"type":   "Internal server error",
					"url":    c.Request().URI(),
					"status": fmt.Sprintf("%d", c.Response().Status()),
					"body":   body,
				}
				raven.SetHttpContext(newHTTPFromCtx(c))
				raven.CaptureError(err, tags)
			}

			return err
		})
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

func newHTTPFromCtx(ctx echo.Context) *raven.Http {
	qs, headers, cookies := getHTTPParams(ctx)

	h := &raven.Http{
		Method:  string(ctx.Request().Method()),
		Cookies: cookies,
		Query:   qs,
		URL:     ctx.Request().URI(),
		Headers: headers,
	}
	return h
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
		l := l.Logger.With(
			zap.String("source", "request"),
		)

		//all except latency to string
		var ip, method, path string
		var header engine.Header
		var status int
		var latency time.Duration
		var startTime, endTime time.Time

		path = c.Path()
		method = c.Request().Method()
		header = c.Request().Header()

		startTime = time.Now()

		err := next(c)

		//no time.Since in order to format it well after
		endTime = time.Now()
		latency = endTime.Sub(startTime)

		status = c.Response().Status()
		ip = c.Request().RemoteAddress()

		route := c.Get("route")
		if route == nil {
			log.D(l, "Route does not have route set in ctx")
			return err
		}

		reqLog := l.With(
			zap.String("route", route.(string)),
			zap.Time("endTime", endTime),
			zap.Int("statusCode", status),
			zap.Duration("latency", latency),
			zap.String("ip", ip),
			zap.String("method", method),
			zap.String("path", path),
			zap.Object("header", header),
		)

		//request failed
		if status > 399 && status < 500 {
			log.W(reqLog, "Request failed.")
			return err
		}

		//request is ok, but server failed
		if status > 499 {
			log.E(reqLog, "Response failed.")
			return err
		}

		//Everything went ok
		if cm := reqLog.Check(zap.InfoLevel, "Request successful."); cm.OK() {
			cm.Write()
		}

		return err
	}
}

//NewNewRelicMiddleware returns the logger middleware
func NewNewRelicMiddleware(app *App, theLogger zap.Logger) *NewRelicMiddleware {
	l := &NewRelicMiddleware{App: app, Logger: theLogger}
	return l
}

//NewRelicMiddleware is responsible for logging to Zap all requests
type NewRelicMiddleware struct {
	App    *App
	Logger zap.Logger
}

// Serve serves the middleware
func (nr *NewRelicMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		method := c.Request().Method()
		route := fmt.Sprintf("%s %s", method, c.Path())
		txn := nr.App.NewRelic.StartTransaction(route, nil, nil)
		c.Set("txn", txn)
		defer func() {
			c.Set("txn", nil)
			txn.End()
		}()

		err := next(c)
		if err != nil {
			txn.NoticeError(err)
			return err
		}

		return nil
	}
}
