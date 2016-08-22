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

type createPlayerPayload struct {
	PublicID string
	Name     string
	Metadata map[string]interface{}
}

type updatePlayerPayload struct {
	Name     string
	Metadata map[string]interface{}
}

func validateUpdatePlayerDispatch(game *models.Game, sourcePlayer *models.Player, player *models.Player, metadata map[string]interface{}, l zap.Logger) bool {
	cl := l.With(
		zap.String("playerUpdateMetadataFieldsHookTriggerWhitelist", game.PlayerUpdateMetadataFieldsHookTriggerWhitelist),
	)

	changedName := player.Name != sourcePlayer.Name
	if changedName {
		cl.Debug("Player name changed")
		return true
	}

	if game.PlayerUpdateMetadataFieldsHookTriggerWhitelist == "" {
		cl.Debug("Player has no metadata whitelist for update hook")
		return true
	}

	cl.Debug("Verifying fields for player update hook dispatch...")
	fields := strings.Split(game.PlayerUpdateMetadataFieldsHookTriggerWhitelist, ",")
	for _, field := range fields {
		oldVal, existsOld := sourcePlayer.Metadata[field]
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
