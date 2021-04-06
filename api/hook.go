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
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

//CreateHookHandler is the handler responsible for creating new hooks
func CreateHookHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "CreateHook")
		start := time.Now()
		gameID := c.Param("gameID")

		db := app.Db(c.StdContext())

		logger := app.Logger.With(
			zap.String("source", "CreateHookHandler"),
			zap.String("operation", "createHook"),
			zap.String("gameID", gameID),
		)

		var payload HookPayload

		if err := LoadJSONPayload(&payload, c, logger); err != nil {
			log.E(logger, "Failed to parse json payload.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}

		log.D(logger, "Creating hook...")
		hook, err := models.CreateHook(
			db,
			gameID,
			payload.Type,
			payload.HookURL,
		)

		if err != nil {
			log.E(logger, "Failed to create the hook.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(logger, "Created hook successfully.", func(cm log.CM) {
			cm.Write(
				zap.String("hookPublicID", hook.PublicID),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})
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

		db := app.Db(c.StdContext())

		logger := app.Logger.With(
			zap.String("source", "RemoveHookHandler"),
			zap.String("operation", "removeHook"),
			zap.String("gameID", gameID),
			zap.String("hookPublicID", publicID),
		)

		log.D(logger, "Removing hook...")
		err := models.RemoveHook(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			log.E(logger, "Failed to remove hook.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(logger, "Hook removed successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})
		return SucceedWith(map[string]interface{}{}, c)
	}
}
