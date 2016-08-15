// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"time"

	"github.com/kataras/iris"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

// Helper methods are located in membership_helpers module.
// This module is only for handlers

// ApplyForMembershipHandler is the handler responsible for applying for new memberships
func ApplyForMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "applyForMembership"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", clanPublicID),
		)

		var payload applyForMembershipPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		optional, err := getMembershipOptionalParameters(app, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l = l.With(
			zap.String("level", payload.Level),
			zap.String("playerPublicID", payload.PlayerPublicID),
		)

		game, err := app.GetGame(gameID)
		if err != nil {
			l.Warn("Could not find game.")
			FailWith(404, err.Error(), c)
			return
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Debug("Applying for membership...")
		membership, err := models.CreateMembership(
			tx,
			game,
			gameID,
			payload.Level,
			payload.PlayerPublicID,
			clanPublicID,
			payload.PlayerPublicID,
			optional.Message,
		)

		if err != nil {
			txErr := app.Rollback(tx, "Membership application failed", l, err)
			if txErr != nil {
				FailWith(500, txErr.Error(), c)
				return
			}

			l.Error("Membership application failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		err = dispatchMembershipHookByID(
			app, tx, models.MembershipApplicationCreatedHook,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.RequestorID, membership.Message,
		)
		if err != nil {
			txErr := app.Rollback(tx, "Membership application failed", l, err)
			if txErr != nil {
				FailWith(500, txErr.Error(), c)
				return
			}

			l.Error("Membership application created dispatch hook failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		err = app.Commit(tx, "Membership application", l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Membership application created successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{}, c)
	}
}

// InviteForMembershipHandler is the handler responsible for creating new memberships
func InviteForMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		gameID := c.Param("gameID")
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "inviteForMembership"),
			zap.String("gameID", gameID),
			zap.String("clanPublicID", clanPublicID),
		)

		var payload inviteForMembershipPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		optional, err := getMembershipOptionalParameters(app, c)
		if err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l = l.With(
			zap.String("level", payload.Level),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		game, err := app.GetGame(gameID)
		if err != nil {
			l.Warn("Could not find game.")
			FailWith(404, err.Error(), c)
			return
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Debug("Inviting for membership...")
		membership, err := models.CreateMembership(
			tx,
			game,
			gameID,
			payload.Level,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
			optional.Message,
		)

		if err != nil {
			txErr := app.Rollback(tx, "Membership invitation failed", l, err)
			if txErr != nil {
				FailWith(500, txErr.Error(), c)
				return
			}

			l.Error("Membership invitation failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		err = dispatchMembershipHookByID(
			app, tx, models.MembershipApplicationCreatedHook,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.RequestorID, membership.Message,
		)
		if err != nil {
			txErr := app.Rollback(tx, "Membership invitation dispatch hook failed", l, err)
			if txErr != nil {
				FailWith(500, txErr.Error(), c)
				return
			}

			l.Error("Membership invitation dispatch hook failed.", zap.Error(err))
			FailWith(500, err.Error(), c)
			return
		}

		err = app.Commit(tx, "Membership invitation", l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Membership invitation created successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{}, c)
	}
}

// ApproveOrDenyMembershipApplicationHandler is the handler responsible for approving or denying a membership invitation
func ApproveOrDenyMembershipApplicationHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
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

		var payload basePayloadWithRequestorAndPlayerPublicIDs
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l = l.With(
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		game, err := app.GetGame(gameID)
		if err != nil {
			l.Warn("Could not find game.")
			FailWith(404, err.Error(), c)
			return
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Approving/Denying membership application failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		l.Debug("Approving/Denying membership application.")
		membership, err := models.ApproveOrDenyMembershipApplication(
			tx,
			game,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
			action,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Approving/Denying membership application failed.", zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}

		l.Debug("Retrieving requestor details.")
		requestor, err := models.GetPlayerByPublicID(tx, gameID, payload.RequestorPublicID)
		if err != nil {
			msg := "Requestor details retrieval failed."
			txErr := rb(err)
			if txErr == nil {
				l.Error(msg, zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("Requestor details retrieved successfully.")

		hookType := models.MembershipApprovedHook
		if action == "deny" {
			hookType = models.MembershipDeniedHook
		}

		err = dispatchApproveDenyMembershipHookByID(
			app, tx, hookType,
			membership.GameID, membership.ClanID, membership.PlayerID,
			requestor.ID, membership.RequestorID, membership.Message,
		)

		if err != nil {
			msg := "Membership approved/denied application dispatch hook failed."
			txErr := rb(err)
			if txErr == nil {
				l.Error(msg, zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}

		err = app.Commit(tx, "Membership application approval/deny", l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Membership application approved/denied successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{}, c)
	}
}

// ApproveOrDenyMembershipInvitationHandler is the handler responsible for approving or denying a membership invitation
func ApproveOrDenyMembershipInvitationHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
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

		var payload approveOrDenyMembershipInvitationPayload
		if err := LoadJSONPayload(&payload, c, l); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l = l.With(
			zap.String("playerPublicID", payload.PlayerPublicID),
		)

		game, err := app.GetGame(gameID)
		if err != nil {
			l.Warn("Could not find game.")
			FailWith(404, err.Error(), c)
			return
		}

		tx, err := app.BeginTrans(l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Approving/Denying membership invitation failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		l.Debug("Approving/Denying membership invitation...")
		membership, err := models.ApproveOrDenyMembershipInvitation(
			tx,
			game,
			gameID,
			payload.PlayerPublicID,
			clanPublicID,
			action,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Membership invitation approval/deny failed.", zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}

		hookType := models.MembershipApprovedHook
		if action == "deny" {
			hookType = models.MembershipDeniedHook
		}

		err = dispatchApproveDenyMembershipHookByID(
			app, tx, hookType,
			membership.GameID, membership.ClanID, membership.PlayerID,
			membership.PlayerID, membership.RequestorID, membership.Message,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Membership invitation approval/deny hook dispatch failed.", zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Membership invitation approved/denied successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		err = app.Commit(tx, "Membership invitation approval/deny", l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

// DeleteMembershipHandler is the handler responsible for deleting a member
func DeleteMembershipHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "deleteMembership"),
			zap.String("clanPublicID", clanPublicID),
		)

		payload, game, status, err := getPayloadAndGame(app, c, l)
		if err != nil {
			FailWith(status, err.Error(), c)
			return
		}

		l = l.With(
			zap.String("gameID", game.PublicID),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		tx, err := app.BeginTrans(l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Deleting membership failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		l.Debug("Deleting membership...")
		err = models.DeleteMembership(
			tx,
			game,
			game.PublicID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Membership delete failed.", zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}

		err = dispatchMembershipHookByPublicID(
			app, tx, models.MembershipLeftHook,
			game.PublicID, clanPublicID, payload.PlayerPublicID,
			payload.RequestorPublicID,
		)
		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Membership deleted hook dispatch failed.", zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Membership deleted successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		err = app.Commit(tx, "Membership invitation approval/deny", l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

// PromoteOrDemoteMembershipHandler is the handler responsible for promoting or demoting a member
func PromoteOrDemoteMembershipHandler(app *App, action string) func(c *iris.Context) {
	return func(c *iris.Context) {
		start := time.Now()
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "promoteOrDemoteMembership"),
			zap.String("clanPublicID", clanPublicID),
			zap.String("action", action),
		)

		payload, game, status, err := getPayloadAndGame(app, c, l)
		if err != nil {
			FailWith(status, err.Error(), c)
			return
		}

		l = l.With(
			zap.String("gameID", game.PublicID),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		tx, err := app.BeginTrans(l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Promoting/Demoting member failed", l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		l.Debug("Promoting/Demoting member...")
		membership, err := models.PromoteOrDemoteMember(
			tx,
			game,
			game.PublicID,
			payload.PlayerPublicID,
			clanPublicID,
			payload.RequestorPublicID,
			action,
		)

		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Member promotion/demotion failed.", zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}
		l.Info("Member promoted/demoted successful.")

		l.Debug("Retrieving promoter/demoter member...")
		requestor, err := models.GetPlayerByPublicID(tx, membership.GameID, payload.RequestorPublicID)
		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Promoter/Demoter member retrieval failed.", zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}
		l.Debug("Promoter/Demoter member retrieved successfully.")

		hookType := models.MembershipPromotedHook
		if action == "demote" {
			hookType = models.MembershipDemotedHook
		}

		err = dispatchMembershipHookByID(
			app, tx, hookType,
			membership.GameID, membership.ClanID, membership.PlayerID,
			requestor.ID, membership.Message,
		)
		if err != nil {
			txErr := rb(err)
			if txErr == nil {
				l.Error("Promote/Demote member hook dispatch failed.", zap.Error(err))
			}
			FailWith(500, err.Error(), c)
			return
		}

		err = app.Commit(tx, "Membership invitation approval/deny", l)
		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		l.Info(
			"Member promoted/demoted successfully.",
			zap.Duration("duration", time.Now().Sub(start)),
		)

		SucceedWith(map[string]interface{}{
			"level": membership.Level,
		}, c)
	}
}
