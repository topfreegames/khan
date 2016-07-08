// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"strings"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/util"
)

// clanPayload maps the payload for the Create Clan route
type clanPayload struct {
	PublicID         string
	Name             string
	OwnerPublicID    string
	Metadata         util.JSON
	AllowApplication bool
	AutoJoin         bool
}

// updateClanPayload maps the payload for the Update Clan route
type updateClanPayload struct {
	Name             string
	OwnerPublicID    string
	Metadata         util.JSON
	AllowApplication bool
	AutoJoin         bool
}

// transferClanOwnershipPayload maps the payload for the Transfer Clan Ownership route
type transferClanOwnershipPayload struct {
	PlayerPublicID string
}

// CreateClanHandler is the handler responsible for creating new clans
func CreateClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")

		var payload clanPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
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
			payload.AllowApplication,
			payload.AutoJoin,
			game.MaxClansPerPlayer,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		result := util.JSON{
			"gameID":           gameID,
			"publicID":         clan.PublicID,
			"name":             clan.Name,
			"membershipCount":  clan.MembershipCount,
			"ownerPublicID":    payload.OwnerPublicID,
			"metadata":         clan.Metadata,
			"allowApplication": clan.AllowApplication,
			"autoJoin":         clan.AutoJoin,
		}
		app.DispatchHooks(gameID, models.ClanCreatedHook, result)

		SucceedWith(util.JSON{
			"publicID": clan.PublicID,
		}, c)
	}
}

// UpdateClanHandler is the handler responsible for updating existing clans
func UpdateClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		var payload updateClanPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		clan, err := models.UpdateClan(
			db,
			gameID,
			publicID,
			payload.Name,
			payload.OwnerPublicID,
			payload.Metadata,
			payload.AllowApplication,
			payload.AutoJoin,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		result := util.JSON{
			"gameID":           gameID,
			"publicID":         clan.PublicID,
			"name":             clan.Name,
			"ownerPublicID":    payload.OwnerPublicID,
			"membershipCount":  clan.MembershipCount,
			"metadata":         clan.Metadata,
			"allowApplication": payload.AllowApplication,
			"autoJoin":         payload.AutoJoin,
		}
		app.DispatchHooks(gameID, models.ClanUpdatedHook, result)

		SucceedWith(util.JSON{}, c)
	}
}

func dispatchClanOwnershipChangeHook(app *App, db models.DB, hookType int, gameID, publicID string) error {
	clan, newOwner, err := models.GetClanAndOwnerByPublicID(db, gameID, publicID)
	if err != nil {
		return err
	}

	newOwnerJSON := newOwner.Serialize()
	delete(newOwnerJSON, "gameID")

	clanJSON := clan.Serialize()
	delete(clanJSON, "gameID")

	result := util.JSON{
		"gameID":   gameID,
		"clan":     clanJSON,
		"newOwner": newOwnerJSON,
	}
	app.DispatchHooks(gameID, hookType, result)

	return nil
}

// LeaveClanHandler is the handler responsible for changing the clan ownership when the owner leaves it
func LeaveClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		db := GetCtxDB(c)

		err := models.LeaveClan(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			if strings.HasPrefix(err.Error(), "Clan was not found with id") {
				FailWith(400, (&models.ModelNotFoundError{Type: "Clan", ID: publicID}).Error(), c)
			} else {
				FailWith(500, err.Error(), c)
			}
			return
		}

		err = dispatchClanOwnershipChangeHook(app, db, models.ClanLeaveHook, gameID, publicID)
		if err != nil {
			FailWith(500, err.Error(), c)
		}

		SucceedWith(util.JSON{}, c)
	}
}

// TransferOwnershipHandler is the handler responsible for transferring the clan ownership to another clan member
func TransferOwnershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		var payload transferClanOwnershipPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
			return
		}

		db := GetCtxDB(c)

		err = models.TransferClanOwnership(
			db,
			gameID,
			publicID,
			payload.PlayerPublicID,
			game.MembershipLevels,
			game.MaxMembershipLevel,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		err = dispatchClanOwnershipChangeHook(app, db, models.ClanTransferOwnershipHook, gameID, publicID)
		if err != nil {
			FailWith(500, err.Error(), c)
		}

		SucceedWith(util.JSON{}, c)
	}
}

// ListClansHandler is the handler responsible for returning a list of all clans
func ListClansHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.Param("gameID")

		clans, err := models.GetAllClans(
			db,
			gameID,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		serializedClans := serializeClans(clans, true)
		SucceedWith(util.JSON{
			"clans": serializedClans,
		}, c)
	}
}

// SearchClansHandler is the handler responsible for searching for clans
func SearchClansHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.Param("gameID")
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
		SucceedWith(util.JSON{
			"clans": serializedClans,
		}, c)
	}
}

// RetrieveClanHandler is the handler responsible for returning details for a given clan
func RetrieveClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		game, err := app.GetGame(gameID)
		if err != nil {
			FailWith(404, err.Error(), c)
		}

		clan, err := models.GetClanDetails(
			db,
			gameID,
			publicID,
			game.MaxClansPerPlayer,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(clan, c)
	}
}

// RetrieveClanSummaryHandler is the handler responsible for returning details summary for a given clan
func RetrieveClanSummaryHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		db := GetCtxDB(c)
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		clan, err := models.GetClanSummary(
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

func serializeClans(clans []models.Clan, includePublicID bool) []util.JSON {
	serializedClans := make([]util.JSON, len(clans))
	for i, clan := range clans {
		serializedClans[i] = serializeClan(&clan, includePublicID)
	}

	return serializedClans
}

func serializeClan(clan *models.Clan, includePublicID bool) util.JSON {
	serial := util.JSON{
		"name":             clan.Name,
		"metadata":         clan.Metadata,
		"allowApplication": clan.AllowApplication,
		"autoJoin":         clan.AutoJoin,
		"membershipCount":  clan.MembershipCount,
	}

	if includePublicID {
		serial["publicID"] = clan.PublicID
	}

	return serial
}
