// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"strings"

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

func dispatchClanOwnershipChangeHook(app *App, db models.DB, hookType int, clan *models.Clan, previousOwner *models.Player, newOwner *models.Player) error {
	newOwnerPublicID := ""
	if newOwner != nil {
		newOwnerPublicID = newOwner.PublicID
	}

	l := app.Logger.With(
		zap.String("source", "clanHandler"),
		zap.String("operation", "dispatchClanOwnershipChangeHook"),
		zap.Int("hookType", hookType),
		zap.String("gameID", clan.GameID),
		zap.String("clanPublicID", clan.PublicID),
		zap.String("newOwnerPublicID", newOwnerPublicID),
		zap.String("previousOwnerPublicID", previousOwner.PublicID),
	)

	previousOwnerJSON := previousOwner.Serialize()
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
		newOwnerJSON := newOwner.Serialize()
		delete(newOwnerJSON, "gameID")
		result["newOwner"] = newOwnerJSON
		result["isDeleted"] = false
	}

	l.Debug("Dispatching hook...")
	app.DispatchHooks(clan.GameID, hookType, result)
	l.Debug("Hook dispatch succeeded.")

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

func validateUpdateClanDispatch(game *models.Game, sourceClan *models.Clan, clan *models.Clan, metadata map[string]interface{}, l zap.Logger) bool {
	cl := l.With(
		zap.String("clanUpdateMetadataFieldsHookTriggerWhitelist", game.ClanUpdateMetadataFieldsHookTriggerWhitelist),
	)

	changedName := clan.Name != sourceClan.Name
	changedAllowApplication := clan.AllowApplication != sourceClan.AllowApplication
	changedAutoJoin := clan.AutoJoin != sourceClan.AutoJoin
	if changedName || changedAllowApplication || changedAutoJoin {
		cl.Debug("One of the main clan properties changed")
		return true
	}

	if game.ClanUpdateMetadataFieldsHookTriggerWhitelist == "" {
		cl.Debug("Clan has no metadata whitelist for update hook")
		return true
	}

	cl.Debug("Verifying fields for clan update hook dispatch...")
	fields := strings.Split(game.ClanUpdateMetadataFieldsHookTriggerWhitelist, ",")
	for _, field := range fields {
		oldVal, existsOld := sourceClan.Metadata[field]
		newVal, existsNew := metadata[field]
		l.Debug(
			"Verifying field for change...",
			zap.Bool("existsOld", existsOld),
			zap.Bool("existsNew", existsNew),
			zap.Object("oldVal", oldVal),
			zap.Object("newVal", newVal),
			zap.String("field", field),
		)
		//fmt.Println("field", field, "existsOld", existsOld, "oldVal", oldVal, "existsNew", existsNew, "newVal", newVal)

		if existsOld != existsNew {
			l.Debug("Found difference in field. Dispatching hook...", zap.String("field", field))
			return true
		}

		if existsOld && oldVal != newVal {
			l.Debug("Found difference in field. Dispatching hook...", zap.String("field", field))
			return true
		}
	}

	return false
}
