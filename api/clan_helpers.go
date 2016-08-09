// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
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
