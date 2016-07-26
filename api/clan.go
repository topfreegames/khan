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
	"github.com/uber-go/zap"
)

// clanPayload maps the payload for the Create Clan route
type clanPayload struct {
	PublicID         string
	Name             string
	OwnerPublicID    string
	Metadata         map[string]interface{}
	AllowApplication bool
	AutoJoin         bool
}

// updateClanPayload maps the payload for the Update Clan route
type updateClanPayload struct {
	Name             string
	OwnerPublicID    string
	Metadata         map[string]interface{}
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

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "createClan"),
			zap.String("gameID", gameID),
		)

		var payload clanPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		game, err := app.GetGame(gameID)
		if err != nil {
			l.Warn("Could not find game.")
			FailWith(404, err.Error(), c)
			return
		}

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Creating clan...")
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
			l.Error("Create clan failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Info("Clan created successfully.")

		result := map[string]interface{}{
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

		SucceedWith(map[string]interface{}{
			"publicID": clan.PublicID,
		}, c)
	}
}

// UpdateClanHandler is the handler responsible for updating existing clans
func UpdateClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "updateClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		var payload updateClanPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Updating clan...")
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
			l.Error("Clan update failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Info("Clan updated successfully.")

		result := map[string]interface{}{
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

		SucceedWith(map[string]interface{}{}, c)
	}
}

func dispatchClanOwnershipChangeHook(app *App, db models.DB, hookType int, gameID, publicID string) error {
	l := app.Logger.With(
		zap.String("source", "clanHandler"),
		zap.String("operation", "dispatchClanOwnershipChangeHook"),
		zap.Int("hookType", hookType),
		zap.String("gameID", gameID),
		zap.String("clanPublicID", publicID),
	)

	l.Debug("Retrieve Clan Owner by PublicID...")
	clan, newOwner, err := models.GetClanAndOwnerByPublicID(db, gameID, publicID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Clan was not found with id") {
			l.Info("Clan was deleted.", zap.Error(err))
			return nil
		}
		return err
	}
	l.Debug("Clan owner retrieval succeeded.")

	newOwnerJSON := newOwner.Serialize()
	delete(newOwnerJSON, "gameID")

	clanJSON := clan.Serialize()
	delete(clanJSON, "gameID")

	result := map[string]interface{}{
		"gameID":   gameID,
		"clan":     clanJSON,
		"newOwner": newOwnerJSON,
	}
	l.Debug("Dispatching hook...")
	app.DispatchHooks(gameID, hookType, result)
	l.Debug("Hook dispatch succeeded.")

	return nil
}

// LeaveClanHandler is the handler responsible for changing the clan ownership when the owner leaves it
func LeaveClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "leaveClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Leaving clan...")
		err = models.LeaveClan(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			if strings.HasPrefix(err.Error(), "Clan was not found with id") {
				l.Warn("Clan was not found.", zap.Error(err))
				FailWith(400, (&models.ModelNotFoundError{Type: "Clan", ID: publicID}).Error(), c)
				return
			}
			l.Error("Clan leave failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("Clan left successfully.")

		err = dispatchClanOwnershipChangeHook(app, db, models.ClanLeftHook, gameID, publicID)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

// TransferOwnershipHandler is the handler responsible for transferring the clan ownership to another clan member
func TransferOwnershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "transferClanOwnership"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		var payload transferClanOwnershipPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l = l.With(
			zap.String("newOwnerPublicID", payload.PlayerPublicID),
		)

		game, err := app.GetGame(gameID)
		if err != nil {
			l.Warn("Could not find game.")
			FailWith(404, err.Error(), c)
			return
		}

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Transferring clan ownership...")
		err = models.TransferClanOwnership(
			db,
			gameID,
			publicID,
			payload.PlayerPublicID,
			game.MembershipLevels,
			game.MaxMembershipLevel,
		)

		if err != nil {
			l.Error("Clan ownership transfer failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		err = dispatchClanOwnershipChangeHook(app, db, models.ClanOwnershipTransferredHook, gameID, publicID)
		if err != nil {
			l.Error("Clan ownership transfer hook dispatch failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		l.Info("Clan ownership transfer completed successfully.")

		SucceedWith(map[string]interface{}{}, c)
	}
}

// ListClansHandler is the handler responsible for returning a list of all clans
func ListClansHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "ListClans"),
			zap.String("gameID", gameID),
		)

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Retrieving all clans...")
		clans, err := models.GetAllClans(
			db,
			gameID,
		)

		if err != nil {
			l.Error("Retrieve all clans failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		l.Info("Retrieve all clans completed successfully.")
		serializedClans := serializeClans(clans, true)
		SucceedWith(map[string]interface{}{
			"clans": serializedClans,
		}, c)
	}
}

// SearchClansHandler is the handler responsible for searching for clans
func SearchClansHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		term := c.URLParam("term")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "SearchClans"),
			zap.String("gameID", gameID),
			zap.String("term", term),
		)

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		if term == "" {
			l.Warn("Clan search failed due to empty term.")
			FailWith(400, (&models.EmptySearchTermError{}).Error(), c)
			return
		}

		l.Debug("Searching clans...")
		clans, err := models.SearchClan(
			db,
			gameID,
			term,
		)

		if err != nil {
			l.Error("Clan search failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		l.Debug("Clan search successful.")
		serializedClans := serializeClans(clans, true)
		SucceedWith(map[string]interface{}{
			"clans": serializedClans,
		}, c)
	}
}

// RetrieveClanHandler is the handler responsible for returning details for a given clan
func RetrieveClanHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		game, err := app.GetGame(gameID)
		if err != nil {
			l.Warn("Could not find game.")
			FailWith(404, err.Error(), c)
			return
		}

		l.Debug("Retrieving clan details...")
		clan, err := models.GetClanDetails(
			db,
			gameID,
			publicID,
			game.MaxClansPerPlayer,
		)

		if err != nil {
			l.Error("Retrieve clan details failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		l.Info("Clan details retrieved successfully.")
		SucceedWith(clan, c)
	}
}

// RetrieveClanSummaryHandler is the handler responsible for returning details summary for a given clan
func RetrieveClanSummaryHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClanSummary"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Retrieving clan summary...")
		clan, err := models.GetClanSummary(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			l.Error("Clan summary retrieval failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		l.Info("Clan summary retrieved successfully.")
		SucceedWith(clan, c)
	}
}

// RetrieveClansSummariesHandler is the handler responsible for returning details summary for a given
// list of clans
func RetrieveClansSummariesHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		gameID := c.Param("gameID")
		publicIDsStr := c.URLParam("clanPublicIds")

		publicIDs := strings.Split(publicIDsStr, ",")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClansSummaries"),
			zap.String("gameID", gameID),
			zap.String("clanPublicIDsStr", publicIDsStr),
		)

		// split of an empty string returns an array with an empty string
		if len(publicIDs) == 1 && publicIDs[0] == "" {
			l.Debug("Empty query string provided.")
			l.Error("Clans summaries retrieval failed, Empty query string provided.")
			FailWith(400, "No clanPublicIds provided", c)
			return
		}

		l.Debug("Getting DB connection...")
		db, err := GetCtxDB(c)
		if err != nil {
			l.Error("Failed to connect to DB.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("DB Connection successful.")

		l.Debug("Retrieving clans summaries...")
		clans, err := models.GetClansSummaries(
			db,
			gameID,
			publicIDs,
		)

		if err != nil {
			if _, ok := err.(*models.CouldNotFindAllClansError); ok {
				l.Error("Clans summaries retrieval failed, 404.", zap.Error(err))
				FailWith(404, err.Error(), c)
			} else {
				l.Error("Clans summaries retrieval failed, 500.", zap.Error(err))
				FailWith(500, err.Error(), c)
			}
			return
		}

		l.Info("Clans summaries retrieved successfully.")
		clansResponse := map[string]interface{}{
			"clans": clans,
		}

		SucceedWith(clansResponse, c)
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
