// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"strings"

	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

func dispatchClanOwnershipChangeHook(app *App, hookType int, clan *models.Clan, previousOwner *models.Player, newOwner *models.Player) error {
	newOwnerPublicID := ""
	if newOwner != nil {
		newOwnerPublicID = newOwner.PublicID
	}

	logger := app.Logger.With(
		zap.String("source", "clanHandler"),
		zap.String("operation", "dispatchClanOwnershipChangeHook"),
		zap.Int("hookType", hookType),
		zap.String("gameID", clan.GameID),
		zap.String("clanPublicID", clan.PublicID),
		zap.String("newOwnerPublicID", newOwnerPublicID),
		zap.String("previousOwnerPublicID", previousOwner.PublicID),
	)

	previousOwnerJSON := previousOwner.Serialize(app.EncryptionKey)
	delete(previousOwnerJSON, "gameID")

	clanJSON := clan.Serialize()
	delete(clanJSON, "gameID")

	result := map[string]interface{}{
		"gameID":        clan.GameID,
		"clan":          clanJSON,
		"previousOwner": previousOwnerJSON,
		"newOwner":      nil,
		"isDeleted":     true,
	}

	if newOwner != nil {
		newOwnerJSON := newOwner.Serialize(app.EncryptionKey)
		delete(newOwnerJSON, "gameID")
		result["newOwner"] = newOwnerJSON
		result["isDeleted"] = false
	}

	log.D(logger, "Dispatching hook...")
	app.DispatchHooks(clan.GameID, hookType, result)
	log.D(logger, "Hook dispatch succeeded.")

	return nil
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

func validateUpdateClanDispatch(game *models.Game, sourceClan *models.Clan, clan *models.Clan, metadata map[string]interface{}, logger zap.Logger) bool {
	cl := logger.With(
		zap.String("clanUpdateMetadataFieldsHookTriggerWhitelist", game.ClanUpdateMetadataFieldsHookTriggerWhitelist),
	)

	changedName := clan.Name != sourceClan.Name
	changedAllowApplication := clan.AllowApplication != sourceClan.AllowApplication
	changedAutoJoin := clan.AutoJoin != sourceClan.AutoJoin
	if changedName || changedAllowApplication || changedAutoJoin {
		log.D(cl, "One of the main clan properties changed")
		return true
	}

	if game.ClanUpdateMetadataFieldsHookTriggerWhitelist == "" {
		log.D(cl, "Clan has no metadata whitelist for update hook")
		return false
	}

	log.D(cl, "Verifying fields for clan update hook dispatch...")
	fields := strings.Split(game.ClanUpdateMetadataFieldsHookTriggerWhitelist, ",")
	for _, field := range fields {
		oldVal, existsOld := sourceClan.Metadata[field]
		newVal, existsNew := metadata[field]
		log.D(logger, "Verifying field for change...", func(cm log.CM) {
			cm.Write(
				zap.Bool("existsOld", existsOld),
				zap.Bool("existsNew", existsNew),
				zap.Object("oldVal", oldVal),
				zap.Object("newVal", newVal),
				zap.String("field", field),
			)
		})

		if existsOld != existsNew {
			log.D(logger, "Found difference in field. Dispatching hook...", func(cm log.CM) {
				cm.Write(zap.String("field", field))
			})
			return true
		}

		if existsOld && oldVal != newVal {
			log.D(logger, "Found difference in field. Dispatching hook...", func(cm log.CM) {
				cm.Write(zap.String("field", field))
			})
			return true
		}
	}

	return false
}
