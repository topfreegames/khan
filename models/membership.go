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

//Membership relates a player to a clan
type Membership struct {
	ID          int    `db:"id"`
	GameID      string `db:"game_id"`
	Level       int    `db:"membership_level"`
	Approved    bool   `db:"approved"`
	Denied      bool   `db:"denied"`
	PlayerID    int    `db:"player_id"`
	ClanID      int    `db:"clan_id"`
	RequestorID int    `db:"requestor_id"`
	CreatedAt   int64  `db:"created_at"`
	UpdatedAt   int64  `db:"updated_at"`
	DeletedAt   int64  `db:"deleted_at"`
}

//PreInsert populates fields before inserting a new clan
func (m *Membership) PreInsert(s gorp.SqlExecutor) error {
	m.CreatedAt = time.Now().UnixNano()
	m.UpdatedAt = m.CreatedAt
	return nil
}

//PreUpdate populates fields before updating a clan
func (m *Membership) PreUpdate(s gorp.SqlExecutor) error {
	m.UpdatedAt = time.Now().UnixNano()
	return nil
}

//GetMembershipByID returns a membership by id
func GetMembershipByID(db DB, id int) (*Membership, error) {
	obj, err := db.Get(Membership{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"Membership", id}
	}
	return obj.(*Membership), nil
}

//GetApprovedMembershipByPlayerPublicID returns a membership for the player with the given publicID
func GetApprovedMembershipByPlayerPublicID(db DB, gameID string, playerPublicID string) (*Membership, error) {
	var membership Membership
	err := db.SelectOne(&membership, "SELECT memberships.* FROM memberships, players WHERE memberships.game_id=$1 AND memberships.approved=true AND memberships.player_id=players.id AND players.public_id=$2", gameID, playerPublicID)
	if err != nil || &membership == nil {
		return nil, &ModelNotFoundError{"Membership", playerPublicID}
	}
	return &membership, nil
}

//CreateMembership creates a new membership
func CreateMembership(db DB, gameID string, level int, playerPublicID string, clanPublicID string, requestorPublicID string) (*Membership, error) {
	minLevelToCreateMembership := 1 // TODO: get this from some config

	player, err := GetPlayerByPublicID(db, gameID, playerPublicID)
	if err != nil {
		return nil, err
	}

	if requestorPublicID == playerPublicID {
		clan, clanErr := GetClanByPublicID(db, gameID, clanPublicID)
		if clanErr != nil {
			return nil, clanErr
		}
		return createMembershipHelper(db, gameID, level, player.ID, clan.ID, player.ID)
	}

	requestorMembership, _ := GetApprovedMembershipByPlayerPublicID(db, gameID, requestorPublicID)
	if requestorMembership == nil {
		clan, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
		}
		return createMembershipHelper(db, gameID, level, player.ID, clan.ID, clan.OwnerID)
	} else if requestorMembership.Level >= minLevelToCreateMembership {
		return createMembershipHelper(db, gameID, level, player.ID, requestorMembership.ClanID, requestorMembership.PlayerID)
	} else {
		return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
	}
}

func createMembershipHelper(db DB, gameID string, level int, playerID int, clanID int, requestorID int) (*Membership, error) {
	membership := &Membership{
		GameID:      gameID,
		ClanID:      clanID,
		PlayerID:    playerID,
		RequestorID: requestorID,
		Level:       level,
		Approved:    false,
		Denied:      false,
	}

	err := db.Insert(membership)
	if err != nil {
		return nil, err
	}
	return membership, nil
}
