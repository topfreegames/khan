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

//CreatePlayerPayload maps the payload for the Create Player route
type CreatePlayerPayload struct {
	GameID   string
	PublicID string
	Name     string
	Metadata string
}

//CreatePlayerHandler is the handler responsible for creating new players
func CreatePlayerHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload CreatePlayerPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		player, err := models.CreatePlayer(
			payload.GameID,
			payload.PublicID,
			payload.Name,
			payload.Metadata,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"id": player.ID,
		}, c)
	}
}

//SetPlayerHandlersGroup configures the routes for all player related routes
func SetPlayerHandlersGroup(app *App) {
	playerHandlersGroup := app.App.Party("/players", func(c *iris.Context) {
		c.Next()
	})

	playerHandlersGroup.Post("", CreatePlayerHandler(app))
}
