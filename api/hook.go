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

	gorp "gopkg.in/gorp.v1"

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
		l := app.Logger.With(
			zap.String("source", "CreateHookHandler"),
			zap.String("operation", "createHook"),
			zap.String("gameID", gameID),
		)

		var payload HookPayload

		err := WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				log.E(l, "Failed to parse json payload.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}

			return nil
		})
		if err != nil {
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}

		var hook *models.Hook
		var tx *gorp.Transaction
		err = WithSegment("hook-create", c, func() error {
			err := WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}

			log.D(l, "Creating hook...")
			hook, err = models.CreateHook(
				tx,
				gameID,
				payload.Type,
				payload.HookURL,
			)

			if err != nil {
				txErr := app.Rollback(tx, "Failed to create hook", c, l, err)
				if txErr != nil {
					return txErr
				}

				log.E(l, "Failed to create the hook.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}

			return nil
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Create hook", c, l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Created hook successfully.", func(cm log.CM) {
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

		l := app.Logger.With(
			zap.String("source", "RemoveHookHandler"),
			zap.String("operation", "removeHook"),
			zap.String("gameID", gameID),
			zap.String("hookPublicID", publicID),
		)

		var tx *gorp.Transaction
		var err error
		err = WithSegment("hook-remove", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}

			log.D(l, "Removing hook...")
			err = models.RemoveHook(
				tx,
				gameID,
				publicID,
			)

			if err != nil {
				txErr := app.Rollback(tx, "Remove hook failed", c, l, err)
				if txErr != nil {
					return txErr
				}

				log.E(l, "Failed to remove hook.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Remove hook", c, l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Hook removed successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})
		return SucceedWith(map[string]interface{}{}, c)
	}
}
