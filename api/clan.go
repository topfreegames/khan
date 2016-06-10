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

//clanPayload maps the payload for the Create Clan route
type clanPayload struct {
	GameID        string
	PublicID      string
	Name          string
	OwnerPublicID string
	Metadata      string
}

//CreateClanHandler is the handler responsible for creating new clans
func CreateClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload clanPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		clan, err := models.CreateClan(
			payload.GameID,
			payload.PublicID,
			payload.Name,
			payload.OwnerPublicID,
			payload.Metadata,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"id": clan.ID,
		}, c)
	}
}

//UpdateClanHandler is the handler responsible for updating existing clans
func UpdateClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		var payload clanPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		_, err := models.UpdateClan(
			payload.GameID,
			payload.PublicID,
			payload.Name,
			payload.OwnerPublicID,
			payload.Metadata,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

//SetClanHandlersGroup configures the routes for all clan related routes
func SetClanHandlersGroup(app *App) {
	clanHandlersGroup := app.App.Party("/clans", func(c *iris.Context) {
		c.Next()
	})

	clanHandlersGroup.Post("", CreateClanHandler(app))
	clanHandlersGroup.Put("", UpdateClanHandler(app))
}
