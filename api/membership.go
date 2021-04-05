// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/topfreegames/extensions/v9/gorp/interfaces"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

// Helper methods are located in membership_helpers module.
// This module is only for handlers

// ApplyForMembershipHandler is the handler responsible for applying for new memberships
func ApplyForMembershipHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "ApplyForMembership")
		start := time.Now()
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "applyForMembership"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", clanPublicID),
		)

		var payload ApplyForMembershipPayload
		var optional *membershipOptionalParams
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			return FailWith(400, err.Error(), c)
		}

		optional, err := getMembershipOptionalParameters(app, c)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("level", payload.Level),
			zap.String("playerPublicID", payload.PlayerPublicID),
		)

		var membership *models.Membership
		game, err := app.GetGame(c.StdContext(), gameID)
		if err != nil {
			log.W(l, "Could not find game.")
			return FailWith(404, err.Error(), c)
		}

		tx, err := app.BeginTrans(c.StdContext(), l)
		if err != nil {
			return FailWithError(err, c)
		}
		log.D(l, "DB Tx begun successful.")

		log.D(l, "Applying for membership...")
		membership, err = models.CreateMembership(
			tx,
			app.EncryptionKey,
			game,
			gameID,
			payload.Level,
			payload.PlayerPublicID,
			clanPublicID,
			payload.PlayerPublicID,
			optional.Message,
		)
		if err != nil {
			txErr := app.Rollback(tx, "Membership application failed", c, l, err)
			if txErr != nil {
				log.E(l, "Could not rollback transaction", func(cm log.CM) {
					cm.Write(zap.Error(txErr))
				})
			}

			log.E(l, "Membership application failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
		}

		err = dispatchMembershipHookByID(
			app, tx, models.MembershipApplicationCreatedHook,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.RequestorID, membership.Message, membership.Level,
		)
		if err != nil {
			txErr := app.Rollback(tx, "Membership application failed", c, l, err)
			if txErr != nil {
				log.E(l, "Could not rollback transaction", func(cm log.CM) {
					cm.Write(zap.Error(txErr))
				})
			}

			log.E(l, "Membership application created dispatch hook failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Membership application", c, l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Membership application created successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(map[string]interface{}{
			"approved": membership.Approved,
		}, c)
	}
}

// InviteForMembershipHandler is the handler responsible for creating new memberships
func InviteForMembershipHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		var payload InviteForMembershipPayload
		var optional *membershipOptionalParams
		var err error
		var game *models.Game
		var membership *models.Membership
		var tx interfaces.Transaction

		c.Set("route", "InviteForMembership")
		start := time.Now()
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "inviteForMembership"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", clanPublicID),
		)

		if err = LoadJSONPayload(&payload, c, l); err != nil {
			return FailWith(400, err.Error(), c)
		}
		optional, err = getMembershipOptionalParameters(app, c)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("level", payload.Level),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		game, err = app.GetGame(c.StdContext(), gameID)
		if err != nil {
			log.W(l, "Could not find game.")
			return FailWith(404, err.Error(), c)
		}

		tx, err = app.BeginTrans(c.StdContext(), l)
		if err != nil {
			return FailWithError(err, c)
		}
		log.D(l, "DB Tx begun successful.")

		log.D(l, "Inviting for membership...")
		membership, err = models.CreateMembership(
			tx,
			app.EncryptionKey,
			game,
			gameID,
			payload.Level,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
			optional.Message,
		)

		if err != nil {
			txErr := app.Rollback(tx, "Membership invitation failed", c, l, err)
			if txErr != nil {
				return FailWith(http.StatusInternalServerError, err.Error(), c)
			}

			log.E(l, "Membership invitation failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
		}

		err = dispatchMembershipHookByID(
			app, tx, models.MembershipApplicationCreatedHook,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.RequestorID, membership.Message, membership.Level,
		)
		if err != nil {
			txErr := app.Rollback(tx, "Membership invitation dispatch hook failed", c, l, err)
			if txErr != nil {
				log.E(l, "Could not rollback transaction", func(cm log.CM) {
					cm.Write(zap.Error(txErr))
				})
			}

			log.E(l, "Membership invitation dispatch hook failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Membership invitation", c, l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Membership invitation created successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(map[string]interface{}{}, c)
	}
}

// ApproveOrDenyMembershipApplicationHandler is the handler responsible for approving or denying a membership invitation
func ApproveOrDenyMembershipApplicationHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "ApproverOrDenyApplication")
		start := time.Now()
		action := c.Param("action")
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "approveOrDenyApplication"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", clanPublicID),
			zap.String("action", action),
		)

		var payload BasePayloadWithRequestorAndPlayerPublicIDs
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		game, err := app.GetGame(c.StdContext(), gameID)
		if err != nil {
			log.W(l, "Could not find game.")
			return FailWith(404, err.Error(), c)
		}

		tx, err := app.BeginTrans(c.StdContext(), l)
		if err != nil {
			return FailWithError(err, c)
		}
		log.D(l, "DB Tx begun successful.")

		rollback := func(err error) error {
			return app.Rollback(tx, "Approving/Denying membership application failed", c, l, err)
		}

		log.D(l, "Approving/Denying membership application.")
		membership, err := models.ApproveOrDenyMembershipApplication(
			tx,
			app.EncryptionKey,
			game,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
			action,
		)

		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				log.E(l, "Could not rollback transaction", func(cm log.CM) {
					cm.Write(zap.Error(txErr))
				})
			}
			return FailWithError(err, c)
		}

		log.D(l, "Retrieving requestor details.")
		requestor, err := models.GetPlayerByPublicID(tx, app.EncryptionKey, gameID, payload.RequestorPublicID)
		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				log.E(l, "Could not rollback transaction", func(cm log.CM) {
					cm.Write(zap.Error(txErr))
				})
			}
			log.E(l, "Requestor details retrieval failed", func(cm log.CM) {
				cm.Write(zap.Error(txErr))
			})

			return FailWithError(err, c)
		}
		log.D(l, "Requestor details retrieved successfully.")

		if err != nil {
			return FailWithError(err, c)
		}

		hookType := models.MembershipApprovedHook
		if action == "deny" {
			hookType = models.MembershipDeniedHook
		}
		err = dispatchApproveDenyMembershipHookByID(
			app, tx, hookType,
			membership.GameID, membership.ClanID, membership.PlayerID,
			requestor.ID, membership.RequestorID, membership.Message, membership.Level,
		)
		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				log.E(l, "Could not rollback transaction", func(cm log.CM) {
					cm.Write(zap.Error(txErr))
				})
			}

			log.E(l, "Membership approved/denied application dispatch hook failed.", func(cm log.CM) {
				cm.Write(zap.Error(txErr))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Membership application approval/deny", c, l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Membership application approved/denied successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(map[string]interface{}{}, c)
	}
}

// ApproveOrDenyMembershipInvitationHandler is the handler responsible for approving or denying a membership invitation
func ApproveOrDenyMembershipInvitationHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		var membership *models.Membership
		var err error
		var tx interfaces.Transaction

		c.Set("route", "ApproveOrDenyInvitation")
		start := time.Now()
		action := c.Param("action")
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "approveOrDenyInvitation"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", clanPublicID),
			zap.String("action", action),
		)

		var payload ApproveOrDenyMembershipInvitationPayload
		err = LoadJSONPayload(&payload, c, l)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("playerPublicID", payload.PlayerPublicID),
		)

		game, err := app.GetGame(c.StdContext(), gameID)
		if err != nil {
			log.W(l, "Could not find game.")
			return FailWith(404, err.Error(), c)
		}

		rollback := func(err error) error {
			txErr := app.Rollback(tx, "Approving/Denying membership invitation failed", c, l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		tx, err = app.BeginTrans(c.StdContext(), l)
		if err != nil {
			return FailWithError(err, c)
		}
		log.D(l, "DB Tx begun successful.")

		log.D(l, "Approving/Denying membership invitation...")
		membership, err = models.ApproveOrDenyMembershipInvitation(
			tx,
			app.EncryptionKey,
			game,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			action,
		)
		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				log.E(l, "Could not rollback transaction", func(cm log.CM) {
					cm.Write(zap.Error(txErr))
				})
			}

			return FailWithError(err, c)
		}

		hookType := models.MembershipApprovedHook
		if action == "deny" {
			hookType = models.MembershipDeniedHook
		}
		err = dispatchApproveDenyMembershipHookByID(
			app, tx, hookType,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.PlayerID, membership.RequestorID, membership.Message, membership.Level,
		)

		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				log.E(l, "Membership invitation approval/deny hook dispatch failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
			}
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Membership invitation approved/denied successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		err = app.Commit(tx, "Membership invitation approval/deny", c, l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{}, c)
	}
}

// DeleteMembershipHandler is the handler responsible for deleting a member
func DeleteMembershipHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "DeleteMembership")
		start := time.Now()
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "deleteMembership"),
			zap.String("clanPublicID", clanPublicID),
		)

		payload, game, status, err := getPayloadAndGame(app, c, l)
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		l = l.With(
			zap.String("gameID", game.PublicID),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		tx, err := app.BeginTrans(c.StdContext(), l)
		if err != nil {
			return FailWithError(err, c)
		}
		log.D(l, "DB Tx began successfully.")

		rollback := func(err error) error {
			txErr := app.Rollback(tx, "Deleting membership failed", c, l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		log.D(l, "Deleting membership...")
		membership, err := models.DeleteMembership(
			tx,
			game,
			game.PublicID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
		)

		if err != nil {
			txErr := rollback(err)
			if txErr == nil {

			}
			log.E(l, "Membership delete failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
		}

		err = dispatchMembershipHookByPublicID(
			app, tx, models.MembershipLeftHook,
			game.PublicID, clanPublicID, payload.PlayerPublicID,
			payload.RequestorPublicID, membership.Level,
		)
		if err != nil {
			txErr := rollback(err)
			if txErr == nil {
				log.E(l, "Could not rollback transaction", func(cm log.CM) {
					cm.Write(zap.Error(txErr))
				})
			}

			log.E(l, "Membership deleted hook dispatch failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Membership deleted successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		err = app.Commit(tx, "Membership invitation approval/deny", c, l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{}, c)
	}
}

// PromoteOrDemoteMembershipHandler is the handler responsible for promoting or demoting a member
func PromoteOrDemoteMembershipHandler(app *App, action string) func(c echo.Context) error {
	return func(c echo.Context) error {

		c.Set("route", "PromoteOrDemoteMember")
		start := time.Now()
		clanPublicID := c.Param("clanPublicID")

		db := app.Db(c.StdContext())

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "promoteOrDemoteMembership"),
			zap.String("clanPublicID", clanPublicID),
			zap.String("action", action),
		)

		payload, game, status, err := getPayloadAndGame(app, c, l)
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		l = l.With(
			zap.String("gameID", game.PublicID),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		log.D(l, "Promoting/Demoting member...")
		membership, err := models.PromoteOrDemoteMember(
			db,
			game,
			game.PublicID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
			action,
		)

		if err != nil {
			log.E(l, "Member promotion/demotion failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
		}
		log.D(l, "Member promoted/demoted successful.")

		log.D(l, "Retrieving promoter/demoter member...")
		requestor, err := models.GetPlayerByPublicID(db, app.EncryptionKey, membership.GameID, payload.RequestorPublicID)
		if err != nil {
			log.E(l, "Promoter/Demoter member retrieval failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWithError(err, c)
		}
		log.D(l, "Promoter/Demoter member retrieved successfully.")

		hookType := models.MembershipPromotedHook
		if action == "demote" {
			hookType = models.MembershipDemotedHook
		}

		err = dispatchMembershipHookByID(
			app, db, hookType,
			membership.GameID, membership.ClanID, membership.PlayerID,
			requestor.ID, membership.Message, membership.Level,
		)
		if err != nil {
			log.E(l, "Promote/Demote member hook dispatch failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Member promoted/demoted successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(map[string]interface{}{
			"level": membership.Level,
		}, c)
	}
}
