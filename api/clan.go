// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

//CreateClanHandler is the handler responsible for creating new clans
func CreateClanHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "CreateClan")
		start := time.Now()
		gameID := c.Param("gameID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "createClan"),
			zap.String("gameID", gameID),
		)

		var payload clanPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			log.E(l, "Failed to parse json payload.", zap.Error(err))
			return FailWith(400, err.Error(), c)
		}

		game, err := app.GetGame(gameID)
		if err != nil {
			log.W(l, "Could not find game.", zap.Error(err))
			return FailWith(404, err.Error(), c)
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		//rollback function
		rb := func(err error) error {
			txErr := app.Rollback(tx, "Creating clan failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

log.D(		l, "Creating clan...")
		clan, err := models.CreateClan(
			tx,
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
			txErr := rb(err)
			if txErr == nil {
				log.E(l, "Create clan failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}

		clanJSON := map[string]interface{}{
			"publicID":         clan.PublicID,
			"name":             clan.Name,
			"membershipCount":  clan.MembershipCount,
			"ownerPublicID":    payload.OwnerPublicID,
			"metadata":         clan.Metadata,
			"allowApplication": clan.AllowApplication,
			"autoJoin":         clan.AutoJoin,
		}

		result := map[string]interface{}{
			"gameID": gameID,
			"clan":   clanJSON,
		}

log.D(		l, "Dispatching hooks")
		err = app.DispatchHooks(gameID, models.ClanCreatedHook, result)
		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(l, "Clan created hook dispatch failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}
log.D(		l, "Hook dispatched successfully.")

		err = app.Commit(tx, "Clan created", l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

log.I(		l, 
			"Clan created successfully.",
			zap.String("clanPublicID", clan.PublicID),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(map[string]interface{}{
			"publicID": clan.PublicID,
		}, c)
	}
}

// UpdateClanHandler is the handler responsible for updating existing clans
func UpdateClanHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "UpdateClan")
		start := time.Now()
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
			log.E(l, "Could not load payload.", zap.Error(err))
			return FailWith(400, err.Error(), c)
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		//rollback function
		rb := func(err error) error {
			txErr := app.Rollback(tx, "Updating clan failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

log.D(		l, "Retrieving game...")
		game, err := models.GetGameByPublicID(tx, gameID)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(l, "Updating clan failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}
log.D(		l, "Game retrieved successfully")

log.D(		l, "Retrieving clan...")
		beforeUpdateClan, err := models.GetClanByPublicID(tx, gameID, publicID)
		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(l, "Updating clan failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}
log.D(		l, "Clan retrieved successfully")

log.D(		l, "Updating clan...")
		clan, err := models.UpdateClan(
			tx,
			gameID,
			publicID,
			payload.Name,
			payload.OwnerPublicID,
			payload.Metadata,
			payload.AllowApplication,
			payload.AutoJoin,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(l, "Updating clan failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}

		clanJSON := map[string]interface{}{
			"publicID":         clan.PublicID,
			"name":             clan.Name,
			"membershipCount":  clan.MembershipCount,
			"ownerPublicID":    payload.OwnerPublicID,
			"metadata":         clan.Metadata,
			"allowApplication": clan.AllowApplication,
			"autoJoin":         clan.AutoJoin,
		}

		result := map[string]interface{}{
			"gameID": gameID,
			"clan":   clanJSON,
		}

		shouldDispatch := validateUpdateClanDispatch(game, beforeUpdateClan, clan, payload.Metadata, l)
		if shouldDispatch {
log.D(			l, "Dispatching clan update hooks...")
			err = app.DispatchHooks(gameID, models.ClanUpdatedHook, result)
			if err != nil {
				txErr := rb(err)
				if txErr == nil {
					log.E(l, "Clan updated hook dispatch failed.", zap.Error(err))
				}
				return FailWith(500, err.Error(), c)
			}
		}

		err = app.Commit(tx, "Clan updated", l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

log.I(		l, 
			"Clan updated successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)
		return SucceedWith(map[string]interface{}{}, c)
	}
}

// LeaveClanHandler is the handler responsible for changing the clan ownership when the owner leaves it
func LeaveClanHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "LeaveClan")
		start := time.Now()
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "leaveClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		//rollback function
		rb := func(err error) error {
			txErr := app.Rollback(tx, "Leaving clan failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

log.D(		l, "Leaving clan...")
		clan, previousOwner, newOwner, err := models.LeaveClan(
			tx,
			gameID,
			publicID,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				if strings.HasPrefix(err.Error(), "Clan was not found with id") {
					log.W(l, "Clan was not found.", zap.Error(err))
					return FailWith(400, (&models.ModelNotFoundError{Type: "Clan", ID: publicID}).Error(), c)
				}
				log.E(l, "Clan leave failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}

		err = dispatchClanOwnershipChangeHook(app, tx, models.ClanLeftHook, clan, previousOwner, newOwner)
		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(l, "Leaving clan hook dispatch failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}

		pOwnerJSON := previousOwner.Serialize()
		delete(pOwnerJSON, "gameID")

		res := map[string]interface{}{
			"previousOwner": pOwnerJSON,
			"newOwner":      nil,
			"isDeleted":     true,
		}

		if newOwner != nil {
			nOwnerJSON := newOwner.Serialize()
			delete(nOwnerJSON, "gameID")
			res["newOwner"] = nOwnerJSON
			res["isDeleted"] = false
		}

		fields := []zap.Field{
			zap.String("clanPublicID", publicID),
			zap.String("previousOwnerPublicID", previousOwner.PublicID),
			zap.Duration("duration", time.Now().Sub(start)),
		}

		if newOwner != nil {
			fields = append(fields, zap.String("newOwnerPublicID", newOwner.PublicID))
		}

		err = app.Commit(tx, "Clan updated", l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

log.I(		l, "Clan left successfully.", fields...)

		return SucceedWith(res, c)
	}
}

// TransferOwnershipHandler is the handler responsible for transferring the clan ownership to another clan member
func TransferOwnershipHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "TransferClanOwnership")
		start := time.Now()
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
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("newOwnerPublicID", payload.PlayerPublicID),
		)

		game, err := app.GetGame(gameID)
		if err != nil {
			log.W(l, "Could not find game.")
			return FailWith(404, err.Error(), c)
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		//rollback function
		rb := func(err error) error {
			txErr := app.Rollback(tx, "Clan ownership transfer failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

log.D(		l, "Transferring clan ownership...")
		clan, previousOwner, newOwner, err := models.TransferClanOwnership(
			tx,
			gameID,
			publicID,
			payload.PlayerPublicID,
			game.MembershipLevels,
			game.MaxMembershipLevel,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(l, "Clan ownership transfer failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}

		err = dispatchClanOwnershipChangeHook(
			app, tx, models.ClanOwnershipTransferredHook,
			clan, previousOwner, newOwner,
		)
		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(l, "Clan ownership transfer hook dispatch failed.", zap.Error(err))
			}
			return FailWith(500, err.Error(), c)
		}

		pOwnerJSON := previousOwner.Serialize()
		delete(pOwnerJSON, "gameID")

		nOwnerJSON := newOwner.Serialize()
		delete(nOwnerJSON, "gameID")

		err = app.Commit(tx, "Clan ownership transfer", l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

log.I(		l, 
			"Clan ownership transfer completed successfully.",
			zap.String("previousOwnerPublicID", previousOwner.PublicID),
			zap.String("newOwnerPublicID", newOwner.PublicID),
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(map[string]interface{}{
			"previousOwner": pOwnerJSON,
			"newOwner":      nOwnerJSON,
		}, c)
	}
}

// ListClansHandler is the handler responsible for returning a list of all clans
func ListClansHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "ListClans")
		start := time.Now()
		gameID := c.Param("gameID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "ListClans"),
			zap.String("gameID", gameID),
		)

log.D(		l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}
log.D(		l, "DB Connection successful.")

log.D(		l, "Retrieving all clans...")
		clans, err := models.GetAllClans(
			db,
			gameID,
		)

		if err != nil {
			log.E(l, "Retrieve all clans failed.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}

		serializedClans := serializeClans(clans, true)

log.I(		l, 
			"Retrieve all clans completed successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(map[string]interface{}{
			"clans": serializedClans,
		}, c)
	}
}

// SearchClansHandler is the handler responsible for searching for clans
func SearchClansHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "SearchClans")
		start := time.Now()
		gameID := c.Param("gameID")
		term := c.QueryParam("term")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "SearchClans"),
			zap.String("gameID", gameID),
			zap.String("term", term),
		)

log.D(		l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}
log.D(		l, "DB Connection successful.")

		if term == "" {
			log.W(l, "Clan search failed due to empty term.")
			return FailWith(400, (&models.EmptySearchTermError{}).Error(), c)
		}

log.D(		l, "Searching clans...")
		clans, err := models.SearchClan(
			db,
			gameID,
			term,
		)

		if err != nil {
			log.E(l, "Clan search failed.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}

		serializedClans := serializeClans(clans, true)

log.I(		l, 
			"Clan search successful.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(map[string]interface{}{
			"clans": serializedClans,
		}, c)
	}
}

// RetrieveClanHandler is the handler responsible for returning details for a given clan
func RetrieveClanHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "RetrieveClan")
		start := time.Now()
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

log.D(		l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}
log.D(		l, "DB Connection successful.")

		game, err := app.GetGame(gameID)
		if err != nil {
			log.W(l, "Could not find game.")
			return FailWith(404, err.Error(), c)
		}

log.D(		l, "Retrieving clan details...")
		clan, err := models.GetClanDetails(
			db,
			gameID,
			publicID,
			game.MaxClansPerPlayer,
		)

		if err != nil {
			log.E(l, "Retrieve clan details failed.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}

log.I(		l, 
			"Clan details retrieved successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(clan, c)
	}
}

// RetrieveClanSummaryHandler is the handler responsible for returning details summary for a given clan
func RetrieveClanSummaryHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "RetrieveClanSummary")
		start := time.Now()
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClanSummary"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

log.D(		l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}
log.D(		l, "DB Connection successful.")

log.D(		l, "Retrieving clan summary...")
		clan, err := models.GetClanSummary(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			log.E(l, "Clan summary retrieval failed.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}

log.I(		l, 
			"Clan summary retrieved successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		return SucceedWith(clan, c)
	}
}

// RetrieveClansSummariesHandler is the handler responsible for returning details summary for a given
// list of clans
func RetrieveClansSummariesHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "RetrieveClansSummaries")
		start := time.Now()
		gameID := c.Param("gameID")
		publicIDsStr := c.QueryParam("clanPublicIds")

		publicIDs := strings.Split(publicIDsStr, ",")

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClansSummaries"),
			zap.String("gameID", gameID),
			zap.String("clanPublicIDsStr", publicIDsStr),
		)

		// split of an empty string returns an array with an empty string
		if len(publicIDs) == 1 && publicIDs[0] == "" {
log.D(			l, "Empty query string provided.")
			log.E(l, "Clans summaries retrieval failed, Empty query string provided.")
			return FailWith(400, "No clanPublicIds provided", c)
		}

log.D(		l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}
log.D(		l, "DB Connection successful.")

log.D(		l, "Retrieving clans summaries...")
		clans, err := models.GetClansSummaries(
			db,
			gameID,
			publicIDs,
		)

		if err != nil {
			if _, ok := err.(*models.CouldNotFindAllClansError); ok {
				log.E(l, "Clans summaries retrieval failed, 404.", zap.Error(err))
				return FailWith(404, err.Error(), c)
			}

			log.E(l, "Clans summaries retrieval failed, 500.", zap.Error(err))
			return FailWith(500, err.Error(), c)
		}

log.I(		l, 
			"Clans summaries retrieved successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		clansResponse := map[string]interface{}{
			"clans": clans,
		}

		return SucceedWith(clansResponse, c)
	}
}
