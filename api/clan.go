// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/topfreegames/extensions/v9/gorp/interfaces"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

func logClanOwnerID(app *App, c echo.Context, gameID, clanPublicID, when, operation string) {
	logger := app.Logger.With(
		zap.String("source", "clanHandler"),
		zap.String("operation", operation),
		zap.String("gameID", gameID),
		zap.String("clanPublicID", clanPublicID),
		zap.String("when", when),
	)

	db, err := app.GetCtxDB(c)
	if err != nil {
		log.E(logger, "Failed to fetch db when logging clan ownerID.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return
	}

	clan, err := models.GetClanByPublicID(db, gameID, clanPublicID)
	if err != nil {
		log.E(logger, "Failed to fetch clan when logging clan ownerID.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return
	}
	if clan == nil {
		log.E(logger, "Clan do not exists when logging clan ownerID.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return
	}

	log.I(logger, "Logged clan ownerID successfully.", func(cm log.CM) {
		cm.Write(
			zap.Int64("ownerID", clan.OwnerID),
		)
	})
}

//CreateClanHandler is the handler responsible for creating new clans
func CreateClanHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "CreateClan")
		start := time.Now()
		gameID := c.Param("gameID")

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "createClan"),
			zap.String("gameID", gameID),
		)

		var payload CreateClanPayload
		if err := LoadJSONPayload(&payload, c, logger); err != nil {
			log.E(logger, "Failed to parse json payload.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}

		game, err := app.GetGame(c.StdContext(), gameID)
		if err != nil {
			log.W(logger, "Could not find game.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(404, err.Error(), c)
		}

		var clan *models.Clan
		var tx interfaces.Transaction

		rollback := func(err error) error {
			txErr := app.Rollback(tx, "Creating clan failed", c, logger, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		tx, err = app.BeginTrans(c.StdContext(), logger)
		if err != nil {
			return FailWithError(err, c)
		}

		log.D(logger, "DB Tx begun successful.")

		log.D(logger, "Creating clan...")
		clan, err = models.CreateClan(
			tx,
			app.EncryptionKey,
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
			txErr := rollback(err)
			if txErr == nil {
				log.E(logger, "Create clan failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
			}
			return FailWithError(err, c)
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

		log.D(logger, "Dispatching hooks")
		err = app.DispatchHooks(gameID, models.ClanCreatedHook, result)
		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				log.E(logger, "Clan created hook dispatch failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
			}
			return FailWith(500, err.Error(), c)
		}
		log.D(logger, "Hook dispatched successfully.")

		err = app.Commit(tx, "Clan created", c, logger)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.D(logger, "Clan created successfully.", func(cm log.CM) {
			cm.Write(
				zap.String("clanPublicID", clan.PublicID),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})

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

		db := app.Db(c.StdContext())

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "updateClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		var payload UpdateClanPayload
		if err := LoadJSONPayload(&payload, c, logger); err != nil {
			log.E(logger, "Could not load payload.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(400, err.Error(), c)
		}

		var clan, beforeUpdateClan *models.Clan

		log.D(logger, "Retrieving game...")
		game, err := models.GetGameByPublicID(db, gameID)
		if err != nil {
			log.E(logger, "Updating clan failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
		}
		log.D(logger, "Game retrieved successfully")

		log.D(logger, "Retrieving clan...")
		beforeUpdateClan, err = models.GetClanByPublicID(db, gameID, publicID)
		if err != nil {
			log.E(logger, "Updating clan failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
		}
		log.D(logger, "Clan retrieved successfully")

		log.D(logger, "Updating clan...")
		clan, err = models.UpdateClan(
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
			log.E(logger, "Updating clan failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
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

		shouldDispatch := validateUpdateClanDispatch(game, beforeUpdateClan, clan, payload.Metadata, logger)
		if shouldDispatch {
			log.D(logger, "Dispatching clan update hooks...")
			err = app.DispatchHooks(gameID, models.ClanUpdatedHook, result)
			if err != nil {
				log.E(logger, "Clan updated hook dispatch failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return FailWith(500, err.Error(), c)
			}
		}

		log.D(logger, "Clan updated successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})
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

		logClanOwnerID(app, c, gameID, publicID, "before", "leaveClan")
		defer logClanOwnerID(app, c, gameID, publicID, "after", "leaveClan")

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "leaveClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		var tx interfaces.Transaction
		var clan *models.Clan
		var previousOwner, newOwner *models.Player
		var err error

		rollback := func(err error) error {
			txErr := app.Rollback(tx, "Leaving clan failed", c, logger, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		tx, err = app.BeginTrans(c.StdContext(), logger)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}
		log.D(logger, "DB Tx begun successful.")

		log.D(logger, "Leaving clan...")
		clan, previousOwner, newOwner, err = models.LeaveClan(
			tx,
			app.EncryptionKey,
			gameID,
			publicID,
		)
		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				if strings.HasPrefix(err.Error(), "Clan was not found with id") {
					log.W(logger, "Clan was not found.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
					notFoundError := &models.ModelNotFoundError{Type: "Clan", ID: publicID}
					return FailWithError(notFoundError, c)
				}
				log.E(logger, "Clan leave failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
			}
			return FailWith(500, err.Error(), c)
		}

		err = dispatchClanOwnershipChangeHook(app, models.ClanLeftHook, clan, previousOwner, newOwner)
		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				log.E(logger, "Leaving clan hook dispatch failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
			}
			return FailWith(500, err.Error(), c)
		}

		res := map[string]interface{}{}
		fields := []zap.Field{}

		previousOwnerJSON := previousOwner.Serialize(app.EncryptionKey)
		delete(previousOwnerJSON, "gameID")

		res["previousOwner"] = previousOwnerJSON
		res["newOwner"] = nil
		res["isDeleted"] = true

		if newOwner != nil {
			newOwnerJSON := newOwner.Serialize(app.EncryptionKey)
			delete(newOwnerJSON, "gameID")
			res["newOwner"] = newOwnerJSON
			res["isDeleted"] = false
		}

		fields = append(fields, zap.String("clanPublicID", publicID))
		fields = append(fields, zap.String("previousOwnerPublicID", previousOwner.PublicID))
		fields = append(fields, zap.Duration("duration", time.Now().Sub(start)))

		if newOwner != nil {
			fields = append(fields, zap.String("newOwnerPublicID", newOwner.PublicID))
		}

		err = app.Commit(tx, "Left clan", c, logger)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.I(logger, "Left clan successfully.", func(cm log.CM) {
			cm.Write(fields...)
		})

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

		logClanOwnerID(app, c, gameID, publicID, "before", "transferClanOwnership")
		defer logClanOwnerID(app, c, gameID, publicID, "after", "transferClanOwnership")

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "transferClanOwnership"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)
		var payload TransferClanOwnershipPayload
		if err := LoadJSONPayload(&payload, c, logger); err != nil {
			return FailWith(400, err.Error(), c)
		}

		logger = logger.With(
			zap.String("newOwnerPublicID", payload.PlayerPublicID),
		)

		game, err := app.GetGame(c.StdContext(), gameID)
		if err != nil {
			log.W(logger, "Could not find game.")
			return FailWith(404, err.Error(), c)
		}

		var tx interfaces.Transaction
		var clan *models.Clan
		var previousOwner, newOwner *models.Player

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Clan ownership transfer failed", c, logger, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		tx, err = app.BeginTrans(c.StdContext(), logger)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.D(logger, "Transferring clan ownership...")
		clan, previousOwner, newOwner, err = models.TransferClanOwnership(
			tx,
			app.EncryptionKey,
			gameID,
			publicID,
			payload.PlayerPublicID,
			game.MembershipLevels,
			game.MaxMembershipLevel,
		)
		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(logger, "Clan ownership transfer failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
			}
			return FailWith(500, err.Error(), c)
		}

		err = dispatchClanOwnershipChangeHook(
			app, models.ClanOwnershipTransferredHook,
			clan, previousOwner, newOwner,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				log.E(logger, "Clan ownership transfer hook dispatch failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
			}
			return FailWith(500, err.Error(), c)
		}

		pOwnerJSON := previousOwner.Serialize(app.EncryptionKey)
		delete(pOwnerJSON, "gameID")

		nOwnerJSON := newOwner.Serialize(app.EncryptionKey)
		delete(nOwnerJSON, "gameID")

		err = app.Commit(tx, "Clan ownership transfer", c, logger)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.I(logger, "Clan ownership transfer completed successfully.", func(cm log.CM) {
			cm.Write(
				zap.String("previousOwnerPublicID", previousOwner.PublicID),
				zap.String("newOwnerPublicID", newOwner.PublicID),
				zap.Duration("duration", time.Now().Sub(start)),
			)
		})
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

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "ListClans"),
			zap.String("gameID", gameID),
		)

		log.D(logger, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(logger, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(logger, "DB Connection successful.")

		log.D(logger, "Retrieving all clans...")
		clans, err := models.GetAllClans(
			db,
			gameID,
		)

		if err != nil {
			log.E(logger, "Retrieve all clans failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}

		serializedClans := serializeClans(clans, true)

		log.D(logger, "Retrieve all clans completed successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(map[string]interface{}{
			"clans": serializedClans,
		}, c)
	}
}

// SearchClansHandler is the handler responsible for searching for clans
func SearchClansHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "MongoSearchClans")
		start := time.Now()
		gameID := c.Param("gameID")
		term := c.QueryParam("term")
		pageSize := app.Config.GetInt64("search.pageSize")

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "SearchClans"),
			zap.String("gameID", gameID),
			zap.String("term", term),
		)

		if term == "" {
			log.W(logger, "Clan search failed due to empty term.")
			return FailWith(400, (&models.EmptySearchTermError{}).Error(), c)
		}

		log.D(logger, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(logger, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(logger, "DB Connection successful.")

		log.D(logger, "Searching clans...")
		clans, err := models.SearchClan(
			db,
			app.MongoDB.WithContext(c.StdContext()),
			gameID,
			term,
			pageSize,
		)

		if err != nil {
			log.E(logger, "Clan search failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}

		serializedClans := serializeClans(clans, true)

		log.D(logger, "Clan search successful.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

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
		shortID := c.QueryParam("shortID")
		maxPendingApplications := c.QueryParam("maxPendingApplications")
		maxPendingInvites := c.QueryParam("maxPendingInvites")
		pendingApplicationsOrder := c.QueryParam("pendingApplicationsOrder")
		pendingInvitesOrder := c.QueryParam("pendingInvitesOrder")

		options := models.NewDefaultGetClanDetailsOptions(app.Config)
		if maxPendingApplications != "" {
			maxApps, err := strconv.ParseUint(maxPendingApplications, 10, 16)
			if err != nil {
				return FailWith(400, err.Error(), c)
			}
			if int(maxApps) > options.MaxPendingApplications {
				return FailWith(400, fmt.Sprintf("Maximum pending applications above allowed (%v).", options.MaxPendingApplications), c)
			}
			options.MaxPendingApplications = int(maxApps)
		}
		if maxPendingInvites != "" {
			maxInvs, err := strconv.ParseUint(maxPendingInvites, 10, 16)
			if err != nil {
				return FailWith(400, err.Error(), c)
			}
			if int(maxInvs) > options.MaxPendingInvites {
				return FailWith(400, fmt.Sprintf("Maximum pending invites above allowed (%v).", options.MaxPendingInvites), c)
			}
			options.MaxPendingInvites = int(maxInvs)
		}
		if pendingApplicationsOrder != "" {
			if !models.IsValidOrder(pendingApplicationsOrder) {
				return FailWith(400, fmt.Sprintf("Pending applications order is invalid (valid orders are %s or %s).", models.Newest, models.Oldest), c)
			}
			options.PendingApplicationsOrder = pendingApplicationsOrder
		}
		if pendingInvitesOrder != "" {
			if !models.IsValidOrder(pendingInvitesOrder) {
				return FailWith(400, fmt.Sprintf("Pending invites order is invalid (valid orders are %s or %s).", models.Newest, models.Oldest), c)
			}
			options.PendingInvitesOrder = pendingInvitesOrder
		}

		db := app.Db(c.StdContext())

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		log.D(logger, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(logger, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(logger, "DB Connection successful.")

		game, err := app.GetGame(c.StdContext(), gameID)
		if err != nil {
			log.W(logger, "Could not find game.")
			return FailWith(404, err.Error(), c)
		}

		var clan *models.Clan
		if shortID == "true" {
			clan, err = models.GetClanByShortPublicID(db, gameID, publicID)
		} else {
			clan, err = models.GetClanByPublicID(db, gameID, publicID)
		}

		if err != nil {
			log.W(logger, "Could not find clan.")
			return FailWith(404, err.Error(), c)
		}

		log.D(logger, "Retrieving clan details...")
		clanResult, err := models.GetClanDetails(
			db,
			app.EncryptionKey,
			gameID,
			clan,
			game.MaxClansPerPlayer,
			options,
		)

		if err != nil {
			log.E(logger, "Retrieve clan details failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}

		log.D(logger, "Clan details retrieved successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})
		return SucceedWith(clanResult, c)
	}
}

// RetrieveClanMembersHandler retrieves only the clan users
func RetrieveClanMembersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "RetrieveClanUsers")
		start := time.Now()
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClanUsers"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		log.D(logger, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(logger, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(logger, "DB Connection successful.")

		log.D(logger, "Retrieving clan players...")
		clanMembers, err := models.GetClanMembers(
			db,
			gameID,
			publicID,
		)
		if err != nil {
			log.E(logger, "Clan playerids retrieval failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}

		log.D(logger, "Clan playerids retrieved successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(clanMembers, c)

	}
}

// RetrieveClanSummaryHandler is the handler responsible for returning details summary for a given clan
func RetrieveClanSummaryHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "RetrieveClanSummary")
		start := time.Now()
		gameID := c.Param("gameID")
		publicID := c.Param("clanPublicID")

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClanSummary"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		log.D(logger, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(logger, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(logger, "DB Connection successful.")

		log.D(logger, "Retrieving clan summary...")
		clanSummary, err := models.GetClanSummary(
			db,
			gameID,
			publicID,
		)

		if err != nil {
			log.E(logger, "Clan summary retrieval failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
		}

		log.D(logger, "Clan summary retrieved successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(clanSummary, c)
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

		logger := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClansSummaries"),
			zap.String("gameID", gameID),
			zap.String("clanPublicIDsStr", publicIDsStr),
		)

		// split of an empty string returns an array with an empty string
		if len(publicIDs) == 1 && publicIDs[0] == "" {
			log.D(logger, "Empty query string provided.")
			log.E(logger, "Clans summaries retrieval failed, Empty query string provided.")
			return FailWith(400, "No clanPublicIds provided", c)
		}

		log.D(logger, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(logger, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(logger, "DB Connection successful.")

		status := 500
		var missingClans []string
		log.D(logger, "Retrieving clans summaries...")
		clansSummaries, err := app.clansSummariesCache.GetClansSummaries(
			db,
			gameID,
			publicIDs,
		)

		if err != nil {
			if _, ok := err.(*models.CouldNotFindAllClansError); ok {
				e := err.(*models.CouldNotFindAllClansError)
				log.W(logger, "Could not find all clans summaries.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				missingClans = e.ClanIDs
			} else {
				log.E(logger, "Clans summaries retrieval failed, 500.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return FailWith(status, err.Error(), c)
			}
		}

		log.D(logger, "Clans summaries retrieved successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		clansResponse := map[string]interface{}{
			"clans": clansSummaries,
		}

		if missingClans != nil && len(missingClans) > 0 {
			clansResponse["missingClans"] = missingClans
		}

		return SucceedWith(clansResponse, c)
	}
}
