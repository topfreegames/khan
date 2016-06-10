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

//Clan identifies uniquely one clan in a given game
type Clan struct {
	ID        int    `db:"id"`
	GameID    string `db:"game_id"`
	PublicID  string `db:"public_id"`
	Name      string `db:"name"`
	OwnerID   int    `db:"owner_id"`
	Metadata  string `db:"metadata"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
	DeletedAt int64  `db:"deleted_at"`
}

//PreInsert populates fields before inserting a new clan
func (c *Clan) PreInsert(s gorp.SqlExecutor) error {
	c.CreatedAt = time.Now().UnixNano()
	c.UpdatedAt = c.CreatedAt
	return nil
}

//PreUpdate populates fields before updating a clan
func (c *Clan) PreUpdate(s gorp.SqlExecutor) error {
	c.UpdatedAt = time.Now().UnixNano()
	return nil
}

//GetClanByID returns a clan by id
func GetClanByID(id int) (*Clan, error) {
	obj, err := db.Get(Clan{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"Clan", id}
	}
	return obj.(*Clan), nil
}

//GetClanByPublicID returns a clan by its public id
func GetClanByPublicID(gameID string, publicID string) (*Clan, error) {
	var clan Clan
	err := db.SelectOne(&clan, "select * from clans where game_id=$1 and public_id=$2", gameID, publicID)
	if err != nil || &clan == nil {
		return nil, &ModelNotFoundError{"Clan", publicID}
	}
	return &clan, nil
}

//GetClanByPublicIDAndOwnerPublicID returns a clan by its public id and the owner public id
func GetClanByPublicIDAndOwnerPublicID(gameID string, publicID string, ownerPublicID string) (*Clan, error) {
	var clan Clan
	err := db.SelectOne(&clan, "SELECT clans.* FROM clans, players WHERE clans.game_id=$1 AND clans.public_id=$2 AND clans.owner_id=players.id AND players.public_id=$3", gameID, publicID, ownerPublicID)
	if err != nil || &clan == nil {
		return nil, &ModelNotFoundError{"Clan", publicID}
	}
	return &clan, nil
}

//CreateClan creates a new clan
func CreateClan(gameID string, publicID string, name string, ownerPublicID string, metadata string) (*Clan, error) {
	player, err := GetPlayerByPublicID(gameID, ownerPublicID)
	if err != nil {
		return nil, err
	}

	clan := &Clan{
		GameID:   gameID,
		PublicID: publicID,
		Name:     name,
		OwnerID:  player.ID,
		Metadata: metadata,
	}

	err = db.Insert(clan)
	if err != nil {
		return nil, err
	}
	return clan, nil
}

//UpdateClan updates an existing clan
func UpdateClan(gameID string, publicID string, name string, ownerPublicID string, metadata string) (*Clan, error) {
	clan, err := GetClanByPublicIDAndOwnerPublicID(gameID, publicID, ownerPublicID)

	if err != nil {
		return nil, err
	}

	clan.Name = name
	clan.Metadata = metadata

	_, err = db.Update(clan)

	if err != nil {
		return nil, err
	}

	return clan, nil
}
