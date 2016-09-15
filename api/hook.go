// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

type hookPayload struct {
	Type    int
	HookURL string
}

//CreateHookHandler is the handler responsible for creating new hooks
func CreateHookHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
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
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
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
				return FailWith(http.StatusInternalServerError, txErr.Error(), c)
			}

			l.Error("Failed to create the hook.", zap.Error(err))
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Create hook", l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		l.Info(
			"Created hook successfully.",
			zap.String("hookPublicID", hook.PublicID),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(map[string]interface{}{
			"publicID": hook.PublicID,
		}, c)
	}
}

// RemoveHookHandler is the handler responsible for removing existing hooks
func RemoveHookHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
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
			return FailWith(http.StatusInternalServerError, err.Error(), c)
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
				return FailWith(http.StatusInternalServerError, txErr.Error(), c)
			}

			l.Error("Failed to remove hook.", zap.Error(err))
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Remove hook", l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		l.Info(
			"Hook removed successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)
		return SucceedWith(map[string]interface{}{}, c)
	}
}
