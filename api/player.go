// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

type createPlayerPayload struct {
	PublicID string
	Name     string
	Metadata string
}

type updatePlayerPayload struct {
	Name     string
	Metadata string
}

// CreatePlayerHandler is the handler responsible for creating new players
func CreatePlayerHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")

		var payload createPlayerPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		player, err := models.CreatePlayer(
			db,
			gameID,
			payload.PublicID,
			payload.Name,
			payload.Metadata,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		result := util.JSON{
			"success":  true,
			"gameID":   gameID,
			"publicID": player.PublicID,
			"name":     player.Name,
			"metadata": player.Metadata,
		}
		app.DispatchHooks(gameID, models.PlayerCreatedHook, result)

		SucceedWith(result, c)
	}
}

// UpdatePlayerHandler is the handler responsible for updating existing
func UpdatePlayerHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		playerPublicID := c.Param("playerPublicID")

		var payload updatePlayerPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		_, err := models.UpdatePlayer(
			db,
			gameID,
			playerPublicID,
			payload.Name,
			payload.Metadata,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		result := util.JSON{
			"success":  true,
			"gameID":   gameID,
			"publicID": playerPublicID,
			"name":     payload.Name,
			"metadata": payload.Metadata,
		}
		app.DispatchHooks(gameID, models.PlayerUpdatedHook, result)

		SucceedWith(util.JSON{}, c)
	}
}

// RetrievePlayerHandler is the handler responsible for returning details for a given player
func RetrievePlayerHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.Param("gameID")
		publicID := c.Param("playerPublicID")

		player, err := models.GetPlayerDetails(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			if err.Error() == fmt.Sprintf("Player was not found with id: %s", publicID) {
				FailWith(404, err.Error(), c)
				return
			}

			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(player, c)
	}
}
