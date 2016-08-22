// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"time"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

type hookPayload struct {
	Type    int
	HookURL string
}

// CreateHookHandler is the handler responsible for creating new hooks
func CreateHookHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		c.Set("route", "CreateHook")
		start := time.Now()
		gameID := c.Param("gameID")
		l := app.Logger.With(
			zap.String("source", "CreateHookHandler"),
			zap.String("operation", "createHook"),
			zap.String("gameID", gameID),
		)

		var payload hookPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			l.Error("Failed to parse json payload.", zap.Error(err))
			FailWith(400, err.Error(), c)
			return
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Debug("Creating hook...")
		hook, err := models.CreateHook(
			tx,
			gameID,
			payload.Type,
			payload.HookURL,
		)

		if err != nil {
			txErr := app.Rollback(tx, "Failed to create hook", l, err)
			if txErr != nil {
				FailWith(500, txErr.Error(), c)
				return
			}

			l.Error("Failed to create the hook.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		err = app.Commit(tx, "Create hook", l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Created hook successfully.",
			zap.String("hookPublicID", hook.PublicID),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{
			"publicID": hook.PublicID,
		}, c)
	}
}

// RemoveHookHandler is the handler responsible for removing existing hooks
func RemoveHookHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		c.Set("route", "RemoveHook")
		start := time.Now()
		gameID := c.Param("gameID")
		publicID := c.Param("publicID")

		l := app.Logger.With(
			zap.String("source", "RemoveHookHandler"),
			zap.String("operation", "removeHook"),
			zap.String("gameID", gameID),
			zap.String("hookPublicID", publicID),
		)

		tx, err := app.BeginTrans(l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Debug("Removing hook...")
		err = models.RemoveHook(
			tx,
			gameID,
			publicID,
		)

		if err != nil {
			txErr := app.Rollback(tx, "Remove hook failed", l, err)
			if txErr != nil {
				FailWith(500, txErr.Error(), c)
				return
			}

			l.Error("Failed to remove hook.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		err = app.Commit(tx, "Remove hook", l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Hook removed successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)
		SucceedWith(map[string]interface{}{}, c)
	}
}
