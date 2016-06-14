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
	DeletedBy   int    `db:"deleted_by"`
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

//GetMembershipByClanAndPlayerPublicID returns a membership for the clan and the player with the given publicIDs
func GetMembershipByClanAndPlayerPublicID(db DB, gameID string, clanPublicID string, playerPublicID string) (*Membership, error) {
	var membership Membership
	err := db.SelectOne(&membership, "SELECT memberships.* FROM memberships, clans, players WHERE memberships.game_id=$1 AND memberships.clan_id=clans.id AND memberships.player_id=players.id AND clans.public_id=$2 AND players.public_id=$3", gameID, clanPublicID, playerPublicID)
	if err != nil || &membership == nil {
		return nil, &ModelNotFoundError{"Membership", playerPublicID}
	}
	return &membership, nil
}

//ApproveOrDenyMembershipInvitation sets Membership.Approved to true or Membership.Denied to true
func ApproveOrDenyMembershipInvitation(db DB, gameID, playerPublicID, clanPublicID, action string) (*Membership, error) {
	membership, err := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, err
	}

	if membership.Approved || membership.Denied {
		return nil, &CannotApproveOrDenyMembershipAlreadyProcessedError{action}
	}

	if membership.PlayerID != membership.RequestorID {
		return approveOrDenyMembershipHelper(db, membership, action)
	}

	// Cannot approve own application
	return nil, &PlayerCannotApproveOrDenyMembershipError{action, playerPublicID, clanPublicID, playerPublicID}
}

//ApproveOrDenyMembershipApplication sets Membership.Approved to true or Membership.Denied to true
func ApproveOrDenyMembershipApplication(db DB, gameID, playerPublicID, clanPublicID, requestorPublicID, action string) (*Membership, error) {
	minLevelToApproveOrDenyMembership := 1 // TODO: get this from some config

	if playerPublicID == requestorPublicID {
		return nil, &PlayerCannotApproveOrDenyMembershipError{action, playerPublicID, clanPublicID, requestorPublicID}
	}

	membership, err := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, err
	}

	if membership.PlayerID != membership.RequestorID {
		return nil, &PlayerCannotApproveOrDenyMembershipError{action, playerPublicID, clanPublicID, requestorPublicID}
	}

	if membership.Approved || membership.Denied {
		return nil, &CannotApproveOrDenyMembershipAlreadyProcessedError{action}
	}

	reqMembership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		_, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotApproveOrDenyMembershipError{action, playerPublicID, clanPublicID, requestorPublicID}
		}
		return approveOrDenyMembershipHelper(db, membership, action)
	} else if reqMembership.Level >= minLevelToApproveOrDenyMembership && reqMembership.Approved == true {
		return approveOrDenyMembershipHelper(db, membership, action)
	} else {
		return nil, &PlayerCannotApproveOrDenyMembershipError{action, playerPublicID, clanPublicID, requestorPublicID}
	}
}

//CreateMembership creates a new membership
func CreateMembership(db DB, gameID string, level int, playerPublicID, clanPublicID, requestorPublicID string) (*Membership, error) {
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

	reqMembership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		clan, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
		}
		return createMembershipHelper(db, gameID, level, player.ID, clan.ID, clan.OwnerID)
	} else if isValidMember(reqMembership) && reqMembership.Level >= minLevelToCreateMembership {
		return createMembershipHelper(db, gameID, level, player.ID, reqMembership.ClanID, reqMembership.PlayerID)
	} else {
		return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
	}
}

//PromoteOrDemoteMember increments or decrements Membership.Level by one
func PromoteOrDemoteMember(db DB, gameID, playerPublicID, clanPublicID, requestorPublicID, action string) (*Membership, error) {
	demote := action == "demote"
	promote := action == "promote"

	minLevel := 0       // TODO: get this from some config
	maxLevel := 1000000 // TODO: get this from some config

	levelOffset := 0
	if promote {
		levelOffset = 1
	}

	if playerPublicID == requestorPublicID {
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
	}

	membership, err := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, err
	}
	if !isValidMember(membership) {
		return nil, &CannotPromoteOrDemoteInvalidMemberError{action}
	}
	if promote && membership.Level >= maxLevel || demote && membership.Level <= minLevel {
		return nil, &CannotPromoteOrDemoteMemberLevelError{action, membership.Level}
	}

	reqMembership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		_, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
		}
		return promoteOrDemoteMemberHelper(db, membership, action)
	} else if isValidMember(reqMembership) && reqMembership.Level > membership.Level+levelOffset {
		return promoteOrDemoteMemberHelper(db, membership, action)
	}
	return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
}

//DeleteMembership soft deletes a membership
func DeleteMembership(db DB, gameID, playerPublicID, clanPublicID, requestorPublicID string) error {
	membership, err := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return err
	}
	if playerPublicID == requestorPublicID {
		return deleteMembershipHelper(db, membership, membership.PlayerID)
	}
	reqMembership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		clan, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return &PlayerCannotPerformMembershipActionError{"delete", playerPublicID, clanPublicID, requestorPublicID}
		}
		return deleteMembershipHelper(db, membership, clan.OwnerID)
	} else if isValidMember(reqMembership) && reqMembership.Level > membership.Level {
		return deleteMembershipHelper(db, membership, reqMembership.PlayerID)
	}
	return &PlayerCannotPerformMembershipActionError{"delete", playerPublicID, clanPublicID, requestorPublicID}
}

func isValidMember(membership *Membership) bool {
	return membership.Approved && !membership.Denied
}

func approveOrDenyMembershipHelper(db DB, membership *Membership, action string) (*Membership, error) {
	if action == "approve" {
		membership.Approved = true
	} else if action == "deny" {
		membership.Denied = true
	} else {
		return nil, &InvalidMembershipActionError{action}
	}
	_, err := db.Update(membership)
	if err != nil {
		return nil, err
	}
	return membership, nil
}

func createMembershipHelper(db DB, gameID string, level, playerID, clanID, requestorID int) (*Membership, error) {
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

func promoteOrDemoteMemberHelper(db DB, membership *Membership, action string) (*Membership, error) {
	// TODO: make it safer by using something like SET column = column + X

	if action == "promote" {
		membership.Level++
	} else if action == "demote" {
		membership.Level--
	} else {
		return nil, &InvalidMembershipActionError{action}
	}
	_, err := db.Update(membership)
	if err != nil {
		return nil, err
	}
	return membership, nil
}

func deleteMembershipHelper(db DB, membership *Membership, deletedBy int) error {
	membership.DeletedAt = time.Now().UnixNano()
	membership.DeletedBy = deletedBy
	membership.Approved = false
	membership.Denied = false
	_, err := db.Update(membership)
	return err
}
