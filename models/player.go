// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"time"

	"gopkg.in/gorp.v1"
)

//Player identifies uniquely one player in a given game
type Player struct {
	ID        int    `db:"id"`
	GameID    string `db:"game_id"`
	PlayerID  string `db:"player_id"`
	Name      string `db:"name"`
	Metadata  string `db:"metadata"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
	DeletedAt int64  `db:"deleted_at"`
}

//PreInsert populates fields before inserting a new player
func (p *Player) PreInsert(s gorp.SqlExecutor) error {
	p.CreatedAt = time.Now().UnixNano()
	p.UpdatedAt = p.CreatedAt
	return nil
}

//PreUpdate populates fields before updating a player
func (p *Player) PreUpdate(s gorp.SqlExecutor) error {
	p.UpdatedAt = time.Now().UnixNano()
	return nil
}

//GetPlayerByID returns a player by id
func GetPlayerByID(id int) (*Player, error) {
	obj, err := db.Get(Player{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"Player", id}
	}
	return obj.(*Player), nil
}

//GetPlayerByPlayerID returns a player by id
func GetPlayerByPlayerID(gameID string, playerID string) (*Player, error) {
	var player Player
	err := db.SelectOne(&player, "select * from players where game_id=$1 and player_id=$2", gameID, playerID)
	if err != nil || &player == nil {
		return nil, &ModelNotFoundError{"Player", playerID}
	}
	return &player, nil
}

//CreatePlayer creates a new player
func CreatePlayer(gameID string, playerID string, name string, metadata string) (*Player, error) {
	player := &Player{
		GameID:   gameID,
		PlayerID: playerID,
		Name:     name,
		Metadata: metadata,
	}
	err := db.Insert(player)
	if err != nil {
		return nil, err
	}
	return player, nil

}
