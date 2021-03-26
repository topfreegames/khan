// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

// EncryptedPlayer identifies uniquely one player in a given game
type EncryptedPlayer struct {
	PlayerID int64 `db:"player_id"`
}
