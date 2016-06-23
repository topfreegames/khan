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

// Membership relates a player to a clan
type Membership struct {
	ID          int    `db:"id"`
	GameID      string `db:"game_id"`
	Level       int    `db:"membership_level"`
	Approved    bool   `db:"approved"`
	Denied      bool   `db:"denied"`
	Banned      bool   `db:"banned"`
	PlayerID    int    `db:"player_id"`
	ClanID      int    `db:"clan_id"`
	RequestorID int    `db:"requestor_id"`
	CreatedAt   int64  `db:"created_at"`
	UpdatedAt   int64  `db:"updated_at"`
	DeletedBy   int    `db:"deleted_by"`
	DeletedAt   int64  `db:"deleted_at"`
}

// PreInsert populates fields before inserting a new clan
func (m *Membership) PreInsert(s gorp.SqlExecutor) error {
	if m.CreatedAt == 0 {
		m.CreatedAt = time.Now().UnixNano() / 1000000
	}
	m.UpdatedAt = m.CreatedAt
	return nil
}

// PreUpdate populates fields before updating a clan
func (m *Membership) PreUpdate(s gorp.SqlExecutor) error {
	m.UpdatedAt = time.Now().UnixNano() / 1000000
	return nil
}

// GetMembershipByID returns a membership by id
func GetMembershipByID(db DB, id int) (*Membership, error) {
	obj, err := db.Get(Membership{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"Membership", id}
	}
	return obj.(*Membership), nil
}

// GetMembershipByClanAndPlayerPublicID returns a membership for the clan and the player with the given publicIDs
func GetMembershipByClanAndPlayerPublicID(db DB, gameID, clanPublicID, playerPublicID string) (*Membership, error) {
	var membership Membership
	err := db.SelectOne(&membership, "SELECT memberships.* FROM memberships, clans, players WHERE memberships.deleted_at=0 AND memberships.game_id=$1 AND memberships.clan_id=clans.id AND memberships.player_id=players.id AND clans.public_id=$2 AND players.public_id=$3", gameID, clanPublicID, playerPublicID)
	if err != nil || &membership == nil {
		return nil, &ModelNotFoundError{"Membership", playerPublicID}
	}
	return &membership, nil
}

// GetDeletedMembershipByClanAndPlayerPublicID returns a deleted membership for the clan and the player with the given publicIDs
func GetDeletedMembershipByClanAndPlayerPublicID(db DB, gameID, clanPublicID, playerPublicID string) (*Membership, error) {
	var membership Membership
	err := db.SelectOne(&membership, "SELECT memberships.* FROM memberships, clans, players WHERE memberships.deleted_at!=0 AND memberships.game_id=$1 AND memberships.clan_id=clans.id AND memberships.player_id=players.id AND clans.public_id=$2 AND players.public_id=$3", gameID, clanPublicID, playerPublicID)
	if err != nil || &membership == nil {
		return nil, &ModelNotFoundError{"Membership", playerPublicID}
	}
	return &membership, nil
}

// GetOldestMemberWithHighestLevel returns the member with highest level that has the oldest creation date
func GetOldestMemberWithHighestLevel(db DB, gameID, clanPublicID string) (*Membership, error) {
	var membership Membership
	err := db.SelectOne(&membership, "SELECT memberships.* FROM memberships, clans WHERE memberships.deleted_at=0 AND memberships.game_id=$1 AND memberships.clan_id=clans.id AND clans.public_id=$2 ORDER BY memberships.membership_level DESC, memberships.created_at ASC LIMIT 1", gameID, clanPublicID)
	if err != nil || &membership == nil {
		return nil, &ClanHasNoMembersError{clanPublicID}
	}
	return &membership, nil
}

// ClanReachedMaxMemberships returns a bool indicating if the clan reached the max number of members (non deleted and approved)
func ClanReachedMaxMemberships(db DB, gameID, clanPublicID string) error {
	var reachedMaxMembers []bool
	_, err := db.Select(&reachedMaxMembers, "SELECT COUNT(memberships.*) >= games.max_members FROM memberships, clans, games WHERE memberships.deleted_at=0 AND memberships.approved=true AND memberships.game_id=$1 AND memberships.clan_id=clans.id AND clans.public_id=$2 AND games.public_id=$1 GROUP BY games.max_members", gameID, clanPublicID)
	if err != nil {
		return err
	}
	if len(reachedMaxMembers) == 0 {
		return nil
	}
	if reachedMaxMembers[0] {
		return &ClanReachedMaxMembersError{clanPublicID}
	}
	return nil
}

// ApproveOrDenyMembershipInvitation sets Membership.Approved to true or Membership.Denied to true
func ApproveOrDenyMembershipInvitation(db DB, gameID, playerPublicID, clanPublicID, action string) (*Membership, error) {
	membership, err := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, err
	}

	if membership.Approved || membership.Denied {
		return nil, &CannotApproveOrDenyMembershipAlreadyProcessedError{action}
	}

	if membership.PlayerID == membership.RequestorID {
		// Cannot approve own application
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, playerPublicID}
	}

	if action == "approve" {
		reachedMaxMembersError := ClanReachedMaxMemberships(db, gameID, clanPublicID)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
	}
	return approveOrDenyMembershipHelper(db, membership, action)
}

// ApproveOrDenyMembershipApplication sets Membership.Approved to true or Membership.Denied to true
func ApproveOrDenyMembershipApplication(db DB, gameID, playerPublicID, clanPublicID, requestorPublicID, action string) (*Membership, error) {
	if playerPublicID == requestorPublicID {
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
	}

	membership, err := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, err
	}

	if membership.PlayerID != membership.RequestorID {
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
	}

	if membership.Approved || membership.Denied {
		return nil, &CannotApproveOrDenyMembershipAlreadyProcessedError{action}
	}

	if action == "approve" {
		reachedMaxMembersError := ClanReachedMaxMemberships(db, gameID, clanPublicID)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
	}

	reqMembership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		_, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
		}
		return approveOrDenyMembershipHelper(db, membership, action)
	}

	game, err := GetGameByPublicID(db, gameID)
	if err != nil {
		return nil, err
	}
	if !reqMembership.Approved || reqMembership.Level < game.MinLevelToAcceptApplication {
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
	}
	return approveOrDenyMembershipHelper(db, membership, action)
}

// CreateMembership creates a new membership
func CreateMembership(db DB, gameID string, level int, playerPublicID, clanPublicID, requestorPublicID string) (*Membership, error) {
	previousMembership := false
	var playerID int

	membership, _ := GetDeletedMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if membership != nil {
		previousMembership = true
		playerID = membership.PlayerID
	} else {
		player, err := GetPlayerByPublicID(db, gameID, playerPublicID)
		if err != nil {
			return nil, err
		}
		playerID = player.ID
	}

	if requestorPublicID == playerPublicID {
		clan, clanErr := GetClanByPublicID(db, gameID, clanPublicID)
		if clanErr != nil {
			return nil, clanErr
		}
		if !clan.AllowApplication {
			return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
		}

		reachedMaxMembersError := ClanReachedMaxMemberships(db, gameID, clanPublicID)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
		if previousMembership {
			return recreateDeletedMembershipHelper(db, membership, level, membership.PlayerID)
		}
		return createMembershipHelper(db, gameID, level, playerID, clan.ID, playerID, clan.AutoJoin)
	}

	reqMembership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		clan, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
		}
		reachedMaxMembersError := ClanReachedMaxMemberships(db, gameID, clanPublicID)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
		if previousMembership {
			return recreateDeletedMembershipHelper(db, membership, level, clan.OwnerID)
		}
		return createMembershipHelper(db, gameID, level, playerID, clan.ID, clan.OwnerID, false)
	}

	reachedMaxMembersError := ClanReachedMaxMemberships(db, gameID, clanPublicID)
	if reachedMaxMembersError != nil {
		return nil, reachedMaxMembersError
	}
	game, err := GetGameByPublicID(db, gameID)
	if err != nil {
		return nil, err
	}

	if isValidMember(reqMembership) && reqMembership.Level >= game.MinLevelToCreateInvitation {
		if previousMembership {
			return recreateDeletedMembershipHelper(db, membership, level, reqMembership.PlayerID)
		}
		return createMembershipHelper(db, gameID, level, playerID, reqMembership.ClanID, reqMembership.PlayerID, false)
	}
	return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
}

// PromoteOrDemoteMember increments or decrements Membership.Level by one
func PromoteOrDemoteMember(db DB, gameID, playerPublicID, clanPublicID, requestorPublicID, action string) (*Membership, error) {
	demote := action == "demote"
	promote := action == "promote"

	game, gameErr := GetGameByPublicID(db, gameID)
	if gameErr != nil {
		return nil, gameErr
	}

	levelOffset := game.MinLevelOffsetToDemoteMember
	if promote {
		levelOffset = game.MinLevelOffsetToPromoteMember
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

	if promote && membership.Level >= game.MaxMembershipLevel || demote && membership.Level <= game.MinMembershipLevel {
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

// DeleteMembership soft deletes a membership
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
	}
	game, gameErr := GetGameByPublicID(db, gameID)
	if gameErr != nil {
		return gameErr
	}
	if isValidMember(reqMembership) && reqMembership.Level >= game.MinLevelToRemoveMember && reqMembership.Level >= membership.Level+game.MinLevelOffsetToRemoveMember {
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

func createMembershipHelper(db DB, gameID string, level, playerID, clanID, requestorID int, approved bool) (*Membership, error) {
	membership := &Membership{
		GameID:      gameID,
		ClanID:      clanID,
		PlayerID:    playerID,
		RequestorID: requestorID,
		Level:       level,
		Approved:    approved,
		Denied:      false,
	}

	err := db.Insert(membership)
	if err != nil {
		return nil, err
	}
	return membership, nil
}

func recreateDeletedMembershipHelper(db DB, membership *Membership, level, requestorID int) (*Membership, error) {
	membership.RequestorID = requestorID
	membership.Level = level
	membership.Approved = false
	membership.Denied = false

	_, err := db.Update(membership)
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
	membership.DeletedAt = time.Now().UnixNano() / 1000000
	membership.DeletedBy = deletedBy
	membership.Approved = false
	membership.Denied = false

	membership.Banned = deletedBy != membership.PlayerID // TODO: Test this

	_, err := db.Update(membership)
	return err
}
