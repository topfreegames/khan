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

		db := GetCtxDB(c)

		clan, err := models.CreateClan(
			db,
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

		db := GetCtxDB(c)

		_, err := models.UpdateClan(
			db,
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

//ListClansHandler is the handler responsible for returning a list of all clans
func ListClansHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.Get("gameID").(string)

		clans, err := models.GetAllClans(
			db,
			gameID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"clans": clans,
		}, c)
	}
}

//SetClanHandlersGroup configures the routes for all clan related routes
func SetClanHandlersGroup(app *App) {
	gameParty := app.App.Party("/games/:gameID", func(c *iris.Context) {
		gameID := c.Param("gameID")
		c.Set("gameID", gameID)
		c.Next()
	})
	clanHandlersGroup := gameParty.Party("/clans", func(c *iris.Context) {
		c.Next()
	})

	clanHandlersGroup.Get("", ListClansHandler(app))
	clanHandlersGroup.Post("", CreateClanHandler(app))
	clanHandlersGroup.Put("", UpdateClanHandler(app))
}
