// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
)

//CreatePlayerPayload maps the payload for the Create Player route
type CreatePlayerPayload struct {
	GameID   string
	PlayerID string
	Name     string
	Metadata string
}

//CreatePlayerHandler is the handler responsible for creating new players
func CreatePlayerHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload CreatePlayerPayload
		if err := c.ReadJSON(&payload); err != nil {
			panic(err.Error())
		}

		player, err := models.CreatePlayer(
			payload.GameID,
			payload.PlayerID,
			payload.Name,
			payload.Metadata,
		)
		fmt.Println(err)

		if err != nil {
			result, _ := json.Marshal(map[string]interface{}{
				"success": false,
				"reason":  err.Error(),
			})
			c.SetStatusCode(500)
			c.Write(string(result))
			return
		}

		result, _ := json.Marshal(map[string]interface{}{
			"success": true,
			"id":      player.ID,
		})

		c.SetStatusCode(200)
		c.Write(string(result))
	}
}

//SetPlayerHandlersGroup configures the routes for all player related routes
func SetPlayerHandlersGroup(app *App) {
	playerHandlersGroup := app.App.Party("/players", func(c *iris.Context) {
		c.Next()
	})

	playerHandlersGroup.Post("/create", CreatePlayerHandler(app))
}
