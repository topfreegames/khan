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
	"github.com/topfreegames/extensions/gorp/interfaces"
	"github.com/topfreegames/khan/log"
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

		var payload CreateClanPayload
		err := WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				log.E(l, "Failed to parse json payload.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var game *models.Game
		err = WithSegment("game-retrieve", c, func() error {
			game, err = app.GetGame(c.StdContext(), gameID)
			if err != nil {
				log.W(l, "Could not find game.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(404, err.Error(), c)
		}

		var clan *models.Clan
		var tx interfaces.Transaction

		//rollback function
		rb := func(err error) error {
			txErr := app.Rollback(tx, "Creating clan failed", c, l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		err = WithSegment("clan-create", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(c.StdContext(), l)
				return err
			})
			if err != nil {
				return err
			}
			log.D(l, "DB Tx begun successful.")

			log.D(l, "Creating clan...")
			clan, err = models.CreateClan(
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
					log.E(l, "Create clan failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
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

		err = WithSegment("hook-dispatch", c, func() error {
			log.D(l, "Dispatching hooks")
			err = app.DispatchHooks(gameID, models.ClanCreatedHook, result)
			if err != nil {
				txErr := rb(err)
				if txErr == nil {
					log.E(l, "Clan created hook dispatch failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			log.D(l, "Hook dispatched successfully.")
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		err = app.Commit(tx, "Clan created", c, l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.D(l, "Clan created successfully.", func(cm log.CM) {
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

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "updateClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		var payload UpdateClanPayload
		err := WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				log.E(l, "Could not load payload.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var clan, beforeUpdateClan *models.Clan
		var game *models.Game

		err = WithSegment("clan-update", c, func() error {
			err = WithSegment("game-retrieve", c, func() error {
				log.D(l, "Retrieving game...")
				game, err = models.GetGameByPublicID(db, gameID)
				return err
			})

			if err != nil {
				log.E(l, "Updating clan failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			log.D(l, "Game retrieved successfully")

			err = WithSegment("clan-retrieve", c, func() error {
				log.D(l, "Retrieving clan...")
				beforeUpdateClan, err = models.GetClanByPublicID(db, gameID, publicID)
				if err != nil {
					log.E(l, "Updating clan failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
					return err
				}
				log.D(l, "Clan retrieved successfully")
				return nil
			})
			if err != nil {
				return err
			}

			err = WithSegment("clan-update-query", c, func() error {
				log.D(l, "Updating clan...")
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
				return err
			})
			if err != nil {
				log.E(l, "Updating clan failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
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

		err = WithSegment("hook-dispatch", c, func() error {
			shouldDispatch := validateUpdateClanDispatch(game, beforeUpdateClan, clan, payload.Metadata, l)
			if shouldDispatch {
				log.D(l, "Dispatching clan update hooks...")
				err = app.DispatchHooks(gameID, models.ClanUpdatedHook, result)
				if err != nil {
					log.E(l, "Clan updated hook dispatch failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
					return err
				}
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.D(l, "Clan updated successfully.", func(cm log.CM) {
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

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "leaveClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		var tx interfaces.Transaction
		var clan *models.Clan
		var previousOwner, newOwner *models.Player
		var err error

		//rollback function
		rb := func(err error) error {
			txErr := app.Rollback(tx, "Leaving clan failed", c, l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		err = WithSegment("clan-leave", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(c.StdContext(), l)
				return err
			})
			if err != nil {
				return err
			}
			log.D(l, "DB Tx begun successful.")

			err = WithSegment("clan-leave-query", c, func() error {
				log.D(l, "Leaving clan...")
				clan, previousOwner, newOwner, err = models.LeaveClan(
					tx,
					gameID,
					publicID,
				)
				return err
			})

			if err != nil {
				txErr := rb(err)
				if txErr == nil {
					if strings.HasPrefix(err.Error(), "Clan was not found with id") {
						log.W(l, "Clan was not found.", func(cm log.CM) {
							cm.Write(zap.Error(err))
						})
						return &models.ModelNotFoundError{Type: "Clan", ID: publicID}
					}
					log.E(l, "Clan leave failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
			return FailWithError(err, c)
		}

		err = WithSegment("hook-dispatch", c, func() error {
			err = dispatchClanOwnershipChangeHook(app, tx, models.ClanLeftHook, clan, previousOwner, newOwner)
			if err != nil {
				txErr := rb(err)
				if txErr == nil {
					log.E(l, "Leaving clan hook dispatch failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		res := map[string]interface{}{}
		fields := []zap.Field{}

		WithSegment("response-serialize", c, func() error {
			pOwnerJSON := previousOwner.Serialize()
			delete(pOwnerJSON, "gameID")

			res["previousOwner"] = pOwnerJSON
			res["newOwner"] = nil
			res["isDeleted"] = true

			if newOwner != nil {
				nOwnerJSON := newOwner.Serialize()
				delete(nOwnerJSON, "gameID")
				res["newOwner"] = nOwnerJSON
				res["isDeleted"] = false
			}

			fields = append(fields, zap.String("clanPublicID", publicID))
			fields = append(fields, zap.String("previousOwnerPublicID", previousOwner.PublicID))
			fields = append(fields, zap.Duration("duration", time.Now().Sub(start)))

			if newOwner != nil {
				fields = append(fields, zap.String("newOwnerPublicID", newOwner.PublicID))
			}
			return nil
		})

		err = app.Commit(tx, "Clan left", c, l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.D(l, "Clan left successfully.", func(cm log.CM) {
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

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "transferClanOwnership"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)
		var payload TransferClanOwnershipPayload
		err := WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("newOwnerPublicID", payload.PlayerPublicID),
		)

		var game *models.Game

		err = WithSegment("game-retrieve", c, func() error {
			game, err = app.GetGame(c.StdContext(), gameID)
			if err != nil {
				log.W(l, "Could not find game.")
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(404, err.Error(), c)
		}

		var tx interfaces.Transaction
		var clan *models.Clan
		var previousOwner, newOwner *models.Player

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Clan ownership transfer failed", c, l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		err = WithSegment("clan-transfer", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(c.StdContext(), l)
				return err
			})
			if err != nil {
				return err
			}

			err = WithSegment("clan-transfer-query", c, func() error {
				log.D(l, "Transferring clan ownership...")
				clan, previousOwner, newOwner, err = models.TransferClanOwnership(
					tx,
					gameID,
					publicID,
					payload.PlayerPublicID,
					game.MembershipLevels,
					game.MaxMembershipLevel,
				)
				return err
			})

			if err != nil {
				txErr := rb(err)
				if txErr == nil {
					log.E(l, "Clan ownership transfer failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		err = WithSegment("hook-dispatch", c, func() error {
			err = dispatchClanOwnershipChangeHook(
				app, tx, models.ClanOwnershipTransferredHook,
				clan, previousOwner, newOwner,
			)
			if err != nil {
				txErr := rb(err)
				if txErr == nil {
					log.E(l, "Clan ownership transfer hook dispatch failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		var pOwnerJSON, nOwnerJSON map[string]interface{}
		err = WithSegment("response-serialize", c, func() error {
			pOwnerJSON = previousOwner.Serialize()
			delete(pOwnerJSON, "gameID")

			nOwnerJSON = newOwner.Serialize()
			delete(nOwnerJSON, "gameID")

			return nil
		})

		err = app.Commit(tx, "Clan ownership transfer", c, l)
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.D(l, "Clan ownership transfer completed successfully.", func(cm log.CM) {
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

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "ListClans"),
			zap.String("gameID", gameID),
		)

		log.D(l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(l, "DB Connection successful.")

		var clans []models.Clan
		err = WithSegment("clan-get-all", c, func() error {
			log.D(l, "Retrieving all clans...")
			clans, err = models.GetAllClans(
				db,
				gameID,
			)

			if err != nil {
				log.E(l, "Retrieve all clans failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		var serializedClans []map[string]interface{}
		err = WithSegment("response-serialize", c, func() error {
			serializedClans = serializeClans(clans, true)
			return nil
		})

		log.D(l, "Retrieve all clans completed successfully.", func(cm log.CM) {
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

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "SearchClans"),
			zap.String("gameID", gameID),
			zap.String("term", term),
		)

		if term == "" {
			log.W(l, "Clan search failed due to empty term.")
			return FailWith(400, (&models.EmptySearchTermError{}).Error(), c)
		}

		var clans []models.Clan
		var err error
		err = WithSegment("clans-search", c, func() error {
			log.D(l, "Searching clans...")
			clans, err = models.SearchClan(
				app.MongoDB.WithContext(c.StdContext()),
				gameID,
				term,
				pageSize,
			)

			if err != nil {
				log.E(l, "Clan search failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		var serializedClans []map[string]interface{}
		WithSegment("response-serialize", c, func() error {
			serializedClans = serializeClans(clans, true)
			return nil
		})

		log.D(l, "Clan search successful.", func(cm log.CM) {
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

		db := app.Db(c.StdContext())

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClan"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		log.D(l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(l, "DB Connection successful.")

		var game *models.Game
		err = WithSegment("game-retrieve", c, func() error {
			game, err = app.GetGame(c.StdContext(), gameID)
			if err != nil {
				log.W(l, "Could not find game.")
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(404, err.Error(), c)
		}

		var clan *models.Clan
		err = WithSegment("clan-retrieve", c, func() error {
			if shortID == "true" {
				clan, err = models.GetClanByShortPublicID(db, gameID, publicID)
			} else {
				clan, err = models.GetClanByPublicID(db, gameID, publicID)
			}
			if err != nil {
				log.W(l, "Could not find clan.")
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(404, err.Error(), c)
		}

		var clanResult map[string]interface{}
		err = WithSegment("clan-retrieve", c, func() error {
			log.D(l, "Retrieving clan details...")
			clanResult, err = models.GetClanDetails(
				db,
				gameID,
				clan,
				game.MaxClansPerPlayer,
			)

			if err != nil {
				log.E(l, "Retrieve clan details failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}

			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.D(l, "Clan details retrieved successfully.", func(cm log.CM) {
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

		l := app.Logger.With(
			zap.String("source", "clanHandler"),
			zap.String("operation", "RetrieveClanUsers"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", publicID),
		)

		log.D(l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(l, "DB Connection successful.")

		var result map[string]interface{}
		err = WithSegment("clan-get-playerids", c, func() error {
			log.D(l, "Retrieving clan players...")
			result, err = models.GetClanMembers(
				db,
				gameID,
				publicID,
			)
			if err != nil {
				log.E(l, "Clan playerids retrieval failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		log.D(l, "Clan playerids retrieved successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(result, c)

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

		log.D(l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(l, "DB Connection successful.")

		var clan map[string]interface{}
		err = WithSegment("clan-get-summary", c, func() error {
			log.D(l, "Retrieving clan summary...")
			clan, err = models.GetClanSummary(
				db,
				gameID,
				publicID,
			)

			if err != nil {
				log.E(l, "Clan summary retrieval failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWithError(err, c)
		}

		log.D(l, "Clan summary retrieved successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

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
			log.D(l, "Empty query string provided.")
			log.E(l, "Clans summaries retrieval failed, Empty query string provided.")
			return FailWith(400, "No clanPublicIds provided", c)
		}

		log.D(l, "Getting DB connection...")
		db, err := app.GetCtxDB(c)
		if err != nil {
			log.E(l, "Failed to connect to DB.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(500, err.Error(), c)
		}
		log.D(l, "DB Connection successful.")

		status := 500
		var clans []map[string]interface{}
		var missingClans []string
		err = WithSegment("clan-get-summaries", c, func() error {
			log.D(l, "Retrieving clans summaries...")
			clans, err = models.GetClansSummaries(
				db,
				gameID,
				publicIDs,
			)

			if err != nil {
				if _, ok := err.(*models.CouldNotFindAllClansError); ok {
					e := err.(*models.CouldNotFindAllClansError)
					log.W(l, "Could not find all clans summaries.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
					missingClans = e.ClanIDs
					return nil
				}

				log.E(l, "Clans summaries retrieval failed, 500.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}
		log.D(l, "Clans summaries retrieved successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		clansResponse := map[string]interface{}{
			"clans": clans,
		}

		if missingClans != nil && len(missingClans) > 0 {
			clansResponse["missingClans"] = missingClans
		}

		return SucceedWith(clansResponse, c)
	}
}
