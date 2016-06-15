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
	PublicID      string
	Name          string
	OwnerPublicID string
	Metadata      string
}

//CreateClanHandler is the handler responsible for creating new clans
func CreateClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Get("gameID").(string)

		var payload clanPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		clan, err := models.CreateClan(
			db,
			gameID,
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
		gameID := c.Get("gameID").(string)
		publicID := c.Param("publicID")

		var payload clanPayload
		if err := c.ReadJSON(&payload); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		_, err := models.UpdateClan(
			db,
			gameID,
			publicID,
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

func serializeClans(clans []models.Clan, includePublicID bool) []map[string]interface{} {
	serializedClans := make([]map[string]interface{}, len(clans))
	for i, clan := range clans {
		serializedClans[i] = serializeClan(&clan, includePublicID)
	}

	return serializedClans
}

func serializeClan(clan *models.Clan, includePublicID bool) map[string]interface{} {
	serial := map[string]interface{}{
		"name":     clan.Name,
		"metadata": clan.Metadata,
	}

	if includePublicID {
		serial["publicID"] = clan.PublicID
	}

	return serial
}

//ListClansHandler is the handler responsible for returning a list of all clans
func ListClansHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.GetString("gameID")

		clans, err := models.GetAllClans(
			db,
			gameID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		serializedClans := serializeClans(clans, true)
		SucceedWith(map[string]interface{}{
			"clans": serializedClans,
		}, c)
	}
}

//SearchClansHandler is the handler responsible for searching for clans
func SearchClansHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.GetString("gameID")
		term := c.URLParam("term")

		if term == "" {
			FailWith(400, (&models.EmptySearchTermError{}).Error(), c)
			return
		}

		clans, err := models.SearchClan(
			db,
			gameID,
			term,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		serializedClans := serializeClans(clans, true)
		SucceedWith(map[string]interface{}{
			"clans": serializedClans,
		}, c)
	}
}

//RetrieveClanHandler is the handler responsible for returning details for a given clan
func RetrieveClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.GetString("gameID")
		publicID := c.Param("publicID")

		clan, err := models.GetClanDetails(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(clan, c)
	}
}

//SetClanHandlersGroup configures the routes for all clan related routes
func SetClanHandlersGroup(app *App, gameParty iris.IParty) {
	clanHandlersGroup := gameParty.Party("/clans", func(c *iris.Context) {
		c.Next()
	})

	clanHandlersGroup.Get("", ListClansHandler(app))
	clanHandlersGroup.Post("", CreateClanHandler(app))
	clanHandlersGroup.Get("/search", SearchClansHandler(app))
	clanHandlersGroup.Get("/clan/:publicID", RetrieveClanHandler(app))
	clanHandlersGroup.Put("/clan/:publicID", UpdateClanHandler(app))
}
