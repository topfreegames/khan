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

	gorp "gopkg.in/gorp.v1"

	"github.com/labstack/echo"
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
		var err error
		err = WithSegment("payload", c, func() error {
			if err = LoadJSONPayload(&payload, c, l); err != nil {
				return err
			}
			optional, err = getMembershipOptionalParameters(app, c)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("level", payload.Level),
			zap.String("playerPublicID", payload.PlayerPublicID),
		)

		var game *models.Game
		var membership *models.Membership
		err = WithSegment("game-retrieve", c, func() error {
			game, err = app.GetGame(gameID)
			if err != nil {
				log.W(l, "Could not find game.")
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(404, err.Error(), c)
		}

		var tx *gorp.Transaction
		err = WithSegment("membership-apply", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}
			log.D(l, "DB Tx begun successful.")

			err = WithSegment("membership-apply-query", c, func() error {
				log.D(l, "Applying for membership...")
				membership, err = models.CreateMembership(
					tx,
					game,
					gameID,
					payload.Level,
					payload.PlayerPublicID,
					clanPublicID,
					payload.PlayerPublicID,
					optional.Message,
				)
				return err
			})

			if err != nil {
				txErr := app.Rollback(tx, "Membership application failed", c, l, err)
				if txErr != nil {
					return txErr
				}

				log.E(l, "Membership application failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = WithSegment("hook-dispatch", c, func() error {
			err = dispatchMembershipHookByID(
				app, tx, models.MembershipApplicationCreatedHook,
				membership.GameID, membership.ClanID, membership.PlayerID,
				membership.RequestorID, membership.Message, membership.Level,
			)
			if err != nil {
				txErr := app.Rollback(tx, "Membership application failed", c, l, err)
				if txErr != nil {
					return txErr
				}

				log.E(l, "Membership application created dispatch hook failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = app.Commit(tx, "Membership application", c, l)
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		log.I(l, "Membership application created successfully.", func(cm log.CM) {
			cm.Write(zap.Duration("duration", time.Now().Sub(start)))
		})

		return SucceedWith(map[string]interface{}{}, c)
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
		var tx *gorp.Transaction

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

		err = WithSegment("payload", c, func() error {
			if err = LoadJSONPayload(&payload, c, l); err != nil {
				return err
			}
			optional, err = getMembershipOptionalParameters(app, c)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("level", payload.Level),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		err = WithSegment("game-retrieve", c, func() error {
			game, err = app.GetGame(gameID)
			if err != nil {
				log.W(l, "Could not find game.")
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(404, err.Error(), c)
		}

		err = WithSegment("membership-invite", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}
			log.D(l, "DB Tx begun successful.")

			err = WithSegment("membership-invite-query", c, func() error {
				log.D(l, "Inviting for membership...")
				membership, err = models.CreateMembership(
					tx,
					game,
					gameID,
					payload.Level,
					payload.PlayerPublicID,
					clanPublicID,
					payload.RequestorPublicID,
					optional.Message,
				)
				return err
			})
			if err != nil {
				txErr := app.Rollback(tx, "Membership invitation failed", c, l, err)
				if txErr != nil {
					return err
				}

				log.E(l, "Membership invitation failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = WithSegment("hook-dispatch", c, func() error {
			err = dispatchMembershipHookByID(
				app, tx, models.MembershipApplicationCreatedHook,
				membership.GameID, membership.ClanID, membership.PlayerID,
				membership.RequestorID, membership.Message, membership.Level,
			)
			if err != nil {
				txErr := app.Rollback(tx, "Membership invitation dispatch hook failed", c, l, err)
				if txErr != nil {
					return txErr
				}

				log.E(l, "Membership invitation dispatch hook failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})
		if err != nil {
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
		var game *models.Game
		var membership *models.Membership
		var requestor *models.Player
		var err error
		var tx *gorp.Transaction

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
		err = WithSegment("payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, l); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		err = WithSegment("game-retrieve", c, func() error {
			game, err = app.GetGame(gameID)
			if err != nil {
				log.W(l, "Could not find game.")
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(404, err.Error(), c)
		}

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Approving/Denying membership application failed", c, l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		err = WithSegment("membership-approve-deny", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}
			log.D(l, "DB Tx begun successful.")

			err = WithSegment("membership-approve-deny-query", c, func() error {
				log.D(l, "Approving/Denying membership application.")
				membership, err = models.ApproveOrDenyMembershipApplication(
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
						log.E(l, "Approving/Denying membership application failed.", func(cm log.CM) {
							cm.Write(zap.Error(err))
						})
					}
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}

			err = WithSegment("player-retrieve", c, func() error {
				log.D(l, "Retrieving requestor details.")
				requestor, err = models.GetPlayerByPublicID(tx, gameID, payload.RequestorPublicID)
				if err != nil {
					msg := "Requestor details retrieval failed."
					txErr := rb(err)
					if txErr == nil {
						log.E(l, msg, func(cm log.CM) {
							cm.Write(zap.Error(err))
						})
					}
					return err
				}
				log.D(l, "Requestor details retrieved successfully.")
				return nil
			})
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = WithSegment("hook-dispatch", c, func() error {
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
				msg := "Membership approved/denied application dispatch hook failed."
				txErr := rb(err)
				if txErr == nil {
					log.E(l, msg, func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
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
		var game *models.Game
		var membership *models.Membership
		var err error
		var tx *gorp.Transaction

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
		err = WithSegment("payload", c, func() error {
			return LoadJSONPayload(&payload, c, l)
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		l = l.With(
			zap.String("playerPublicID", payload.PlayerPublicID),
		)

		err = WithSegment("game-retrieve", c, func() error {
			game, err = app.GetGame(gameID)
			if err != nil {
				log.W(l, "Could not find game.")
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(404, err.Error(), c)
		}

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Approving/Denying membership invitation failed", c, l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		err = WithSegment("membership-approve-deny", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}
			log.D(l, "DB Tx begun successful.")

			return WithSegment("memership-approve-deny-query", c, func() error {
				log.D(l, "Approving/Denying membership invitation...")
				membership, err = models.ApproveOrDenyMembershipInvitation(
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
						log.E(l, "Membership invitation approval/deny failed.", func(cm log.CM) {
							cm.Write(zap.Error(err))
						})
					}
					return err
				}
				return nil
			})
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = WithSegment("hook-dispatch", c, func() error {
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
				txErr := rb(err)
				if txErr == nil {
					log.E(l, "Membership invitation approval/deny hook dispatch failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
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
		var err error
		var status int
		var payload *BasePayloadWithRequestorAndPlayerPublicIDs
		var game *models.Game
		var membership *models.Membership
		var tx *gorp.Transaction

		c.Set("route", "DeleteMembership")
		start := time.Now()
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "deleteMembership"),
			zap.String("clanPublicID", clanPublicID),
		)

		err = WithSegment("payload", c, func() error {
			payload, game, status, err = getPayloadAndGame(app, c, l)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		l = l.With(
			zap.String("gameID", game.PublicID),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		rb := func(err error) error {
			txErr := app.Rollback(tx, "Deleting membership failed", c, l, err)
			if txErr != nil {
				return txErr
			}

			return nil
		}

		err = WithSegment("membership-delete", c, func() error {
			err = WithSegment("tx-begin", c, func() error {
				tx, err = app.BeginTrans(l)
				return err
			})
			if err != nil {
				return err
			}
			log.D(l, "DB Tx begun successful.")

			log.D(l, "Deleting membership...")
			membership, err = models.DeleteMembership(
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
					log.E(l, "Membership delete failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = WithSegment("hook-dispatch", c, func() error {
			err = dispatchMembershipHookByPublicID(
				app, tx, models.MembershipLeftHook,
				game.PublicID, clanPublicID, payload.PlayerPublicID,
				payload.RequestorPublicID, membership.Level,
			)
			if err != nil {
				txErr := rb(err)
				if txErr == nil {
					log.E(l, "Membership deleted hook dispatch failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
				}
				return err
			}
			return nil
		})
		if err != nil {
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
		var payload *BasePayloadWithRequestorAndPlayerPublicIDs
		var game *models.Game
		var membership *models.Membership
		var requestor *models.Player
		var status int
		var err error

		c.Set("route", "PromoteOrDemoteMember")
		start := time.Now()
		clanPublicID := c.Param("clanPublicID")

		l := app.Logger.With(
			zap.String("source", "membershipHandler"),
			zap.String("operation", "promoteOrDemoteMembership"),
			zap.String("clanPublicID", clanPublicID),
			zap.String("action", action),
		)

		err = WithSegment("payload", c, func() error {
			payload, game, status, err = getPayloadAndGame(app, c, l)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		l = l.With(
			zap.String("gameID", game.PublicID),
			zap.String("playerPublicID", payload.PlayerPublicID),
			zap.String("requestorPublicID", payload.RequestorPublicID),
		)

		err = WithSegment("membership-promote-demote", c, func() error {
			err = WithSegment("membership-promote-demote-query", c, func() error {
				log.D(l, "Promoting/Demoting member...")
				membership, err = models.PromoteOrDemoteMember(
					app.Db,
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
					return err
				}
				log.I(l, "Member promoted/demoted successful.")
				return nil
			})

			if err != nil {
				return err
			}

			err = WithSegment("player-retrieve", c, func() error {
				log.D(l, "Retrieving promoter/demoter member...")
				requestor, err = models.GetPlayerByPublicID(app.Db, membership.GameID, payload.RequestorPublicID)
				if err != nil {
					log.E(l, "Promoter/Demoter member retrieval failed.", func(cm log.CM) {
						cm.Write(zap.Error(err))
					})
					return err
				}
				return nil
			})
			log.D(l, "Promoter/Demoter member retrieved successfully.")
			return err
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		err = WithSegment("hook-dispatch", c, func() error {
			hookType := models.MembershipPromotedHook
			if action == "demote" {
				hookType = models.MembershipDemotedHook
			}

			err = dispatchMembershipHookByID(
				app, app.Db, hookType,
				membership.GameID, membership.ClanID, membership.PlayerID,
				requestor.ID, membership.Message, membership.Level,
			)
			if err != nil {
				log.E(l, "Promote/Demote member hook dispatch failed.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return err
			}
			return nil
		})

		if err != nil {
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
