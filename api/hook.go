// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
)

type hookPayload struct {
	Type    int
	HookURL string
}

// CreateHookHandler is the handler responsible for creating new hooks
func CreateHookHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")

		var payload hookPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db, err := GetCtxDB(c)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		hook, err := models.CreateHook(
			db,
			gameID,
			payload.Type,
			payload.HookURL,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": hook.PublicID,
		}, c)
	}
}

// RemoveHookHandler is the handler responsible for removing existing hooks
func RemoveHookHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("publicID")

		db, err := GetCtxDB(c)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		err = models.RemoveHook(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}
