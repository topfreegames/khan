// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"runtime/debug"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
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
	}
}

// GetCtxDB returns the proper database connection depending on the request context
func GetCtxDB(ctx *iris.Context) models.DB {
	return ctx.Get("db").(models.DB)
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
	OnError func(interface{}, []byte)
}

//Serve executes on error handler when errors happen
func (r RecoveryMiddleware) Serve(ctx *iris.Context) {
	defer func() {
		if err := recover(); err != nil {
			if r.OnError != nil {
				r.OnError(err, debug.Stack())
			}
			ctx.Panic()
		}
	}()
	ctx.Next()
}
