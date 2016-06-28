// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"time"

	"github.com/topfreegames/khan/util"

	"gopkg.in/gorp.v1"
)

// Membership relates a player to a clan
type Membership struct {
	ID          int    `db:"id"`
	GameID      string `db:"game_id"`
	Level       string `db:"membership_level"`
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
	query := `
	SELECT
		m.*
	FROM memberships m
		INNER JOIN clans c ON c.public_id=$1 AND c.id=m.clan_id
		INNER JOIN players p ON p.public_id=$2 AND p.id=m.player_id
	WHERE
		m.game_id=$3 AND
		m.deleted_at=0`

	err := db.SelectOne(&membership, query, clanPublicID, playerPublicID, gameID)
	if err != nil || &membership == nil {
		return nil, &ModelNotFoundError{"Membership", playerPublicID}
	}
	return &membership, nil
}

// GetDeletedMembershipByClanAndPlayerPublicID returns a deleted membership for the clan and the player with the given publicIDs
func GetDeletedMembershipByClanAndPlayerPublicID(db DB, gameID, clanPublicID, playerPublicID string) (*Membership, error) {
	var membership Membership
	query := `
	SELECT
		m.*
	FROM memberships m
		INNER JOIN clans c ON c.public_id=$1 AND c.id=m.clan_id
		INNER JOIN players p ON p.public_id=$2 AND p.id=m.player_id
	WHERE
		m.deleted_at!=0 AND m.game_id=$3`

	err := db.SelectOne(&membership, query, clanPublicID, playerPublicID, gameID)
	if err != nil || &membership == nil {
		return nil, &ModelNotFoundError{"Membership", playerPublicID}
	}
	return &membership, nil
}

// GetOldestMemberWithHighestLevel returns the member with highest level that has the oldest creation date
func GetOldestMemberWithHighestLevel(db DB, gameID, clanPublicID string) (*Membership, error) {
	var membership Membership
	query := `
	SELECT
	 m.*
	FROM memberships m
	 INNER JOIN games g ON g.public_id=$1 AND g.public_id=m.game_id
	 INNER JOIN clans c ON c.public_id=$2 AND c.id=m.clan_id
	ORDER BY
	 g.membership_levels::json->>m.membership_level DESC,
	 m.created_at ASC
	LIMIT 1`
	err := db.SelectOne(&membership, query, gameID, clanPublicID)
	if err != nil || &membership == nil {
		return nil, &ClanHasNoMembersError{clanPublicID}
	}
	return &membership, nil
}

func clanReachedMaxMemberships(db DB, game *Game, clan *Clan, clanID int) error {
	var err error
	if clan == nil {
		clan, err = GetClanByID(db, clanID)
		if err != nil {
			return err
		}
	}
	if clan.MembershipCount >= game.MaxMembers {
		return &ClanReachedMaxMembersError{clan.PublicID}
	}
	return nil
}

// ApproveOrDenyMembershipInvitation sets Membership.Approved to true or Membership.Denied to true
func ApproveOrDenyMembershipInvitation(db DB, game *Game, gameID, playerPublicID, clanPublicID, action string) (*Membership, error) {
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
		player, err := GetPlayerByID(db, membership.PlayerID)
		if err != nil {
			return nil, err
		}
		if player.MembershipCount+player.OwnershipCount >= game.MaxClansPerPlayer {
			return nil, &PlayerReachedMaxClansError{playerPublicID}
		}
		reachedMaxMembersError := clanReachedMaxMemberships(db, game, nil, membership.ClanID)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
	}
	return approveOrDenyMembershipHelper(db, membership, action)
}

// ApproveOrDenyMembershipApplication sets Membership.Approved to true or Membership.Denied to true
func ApproveOrDenyMembershipApplication(db DB, game *Game, gameID, playerPublicID, clanPublicID, requestorPublicID, action string) (*Membership, error) {
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
		player, err := GetPlayerByID(db, membership.PlayerID)
		if err != nil {
			return nil, err
		}
		if player.MembershipCount+player.OwnershipCount >= game.MaxClansPerPlayer {
			return nil, &PlayerReachedMaxClansError{playerPublicID}
		}
		reachedMaxMembersError := clanReachedMaxMemberships(db, game, nil, membership.ClanID)
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

	levelInt := GetLevelIntByLevel(reqMembership.Level, game.MembershipLevels)
	if !reqMembership.Approved || levelInt < game.MinLevelToAcceptApplication {
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
	}
	return approveOrDenyMembershipHelper(db, membership, action)
}

// CreateMembership creates a new membership
func CreateMembership(db DB, game *Game, gameID, level, playerPublicID, clanPublicID, requestorPublicID string) (*Membership, error) {
	previousMembership := false
	var playerID int
	if _, levelValid := game.MembershipLevels[level]; !levelValid {
		return nil, &InvalidLevelForGameError{gameID, level}
	}

	membership, _ := GetDeletedMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if membership != nil {
		previousMembership = true
		playerID = membership.PlayerID
		player, err := GetPlayerByID(db, playerID)
		if err != nil {
			return nil, err
		}
		if player.MembershipCount+player.OwnershipCount >= game.MaxClansPerPlayer {
			return nil, &PlayerReachedMaxClansError{playerPublicID}
		}
	} else {
		player, err := GetPlayerByPublicID(db, gameID, playerPublicID)
		if err != nil {
			return nil, err
		}
		if player.MembershipCount+player.OwnershipCount >= game.MaxClansPerPlayer {
			return nil, &PlayerReachedMaxClansError{playerPublicID}
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

		reachedMaxMembersError := clanReachedMaxMemberships(db, game, clan, -1)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
		if previousMembership {
			return recreateDeletedMembershipHelper(db, membership, level, membership.PlayerID, clan.AutoJoin)
		}
		return createMembershipHelper(db, gameID, level, playerID, clan.ID, playerID, clan.AutoJoin)
	}

	reqMembership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		clan, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
		}
		reachedMaxMembersError := clanReachedMaxMemberships(db, game, clan, -1)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
		if previousMembership {
			return recreateDeletedMembershipHelper(db, membership, level, clan.OwnerID, false)
		}
		return createMembershipHelper(db, gameID, level, playerID, clan.ID, clan.OwnerID, false)
	}

	reachedMaxMembersError := clanReachedMaxMemberships(db, game, nil, reqMembership.ClanID)
	if reachedMaxMembersError != nil {
		return nil, reachedMaxMembersError
	}

	levelInt := GetLevelIntByLevel(reqMembership.Level, game.MembershipLevels)

	if isValidMember(reqMembership) && levelInt >= game.MinLevelToCreateInvitation {
		if previousMembership {
			return recreateDeletedMembershipHelper(db, membership, level, reqMembership.PlayerID, false)
		}
		return createMembershipHelper(db, gameID, level, playerID, reqMembership.ClanID, reqMembership.PlayerID, false)
	}
	return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clanPublicID}
}

// PromoteOrDemoteMember increments or decrements Membership.LevelInt by one
func PromoteOrDemoteMember(db DB, game *Game, gameID, playerPublicID, clanPublicID, requestorPublicID, action string) (*Membership, error) {
	demote := action == "demote"
	promote := action == "promote"

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

	levelInt := GetLevelIntByLevel(membership.Level, game.MembershipLevels)
	if promote && levelInt >= game.MaxMembershipLevel || demote && levelInt <= game.MinMembershipLevel {
		return nil, &CannotPromoteOrDemoteMemberLevelError{action, levelInt}
	}

	reqMembership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		_, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
		}
		return promoteOrDemoteMemberHelper(db, membership, action, game.MembershipLevels)
	}

	reqLevelInt := GetLevelIntByLevel(reqMembership.Level, game.MembershipLevels)
	if isValidMember(reqMembership) && reqLevelInt >= levelInt+levelOffset {
		return promoteOrDemoteMemberHelper(db, membership, action, game.MembershipLevels)
	}
	return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
}

// DeleteMembership soft deletes a membership
func DeleteMembership(db DB, game *Game, gameID, playerPublicID, clanPublicID, requestorPublicID string) error {
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

	levelInt := GetLevelIntByLevel(membership.Level, game.MembershipLevels)
	reqLevelInt := GetLevelIntByLevel(reqMembership.Level, game.MembershipLevels)
	if isValidMember(reqMembership) && reqLevelInt >= game.MinLevelToRemoveMember && reqLevelInt >= levelInt+game.MinLevelOffsetToRemoveMember {
		return deleteMembershipHelper(db, membership, reqMembership.PlayerID)
	}
	return &PlayerCannotPerformMembershipActionError{"delete", playerPublicID, clanPublicID, requestorPublicID}
}

func isValidMember(membership *Membership) bool {
	return membership.Approved && !membership.Denied
}

func approveOrDenyMembershipHelper(db DB, membership *Membership, action string) (*Membership, error) {
	approve := action == "approve"
	if approve {
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
	if approve {
		err = IncrementPlayerMembershipCount(db, membership.PlayerID, 1)
		if err != nil {
			return nil, err
		}
		err = IncrementClanMembershipCount(db, membership.ClanID, 1)
		if err != nil {
			return nil, err
		}
	}
	return membership, nil
}

func createMembershipHelper(db DB, gameID, level string, playerID, clanID, requestorID int, approved bool) (*Membership, error) {
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
	if approved {
		err = IncrementPlayerMembershipCount(db, membership.PlayerID, 1)
		if err != nil {
			return nil, err
		}
		err = IncrementClanMembershipCount(db, membership.ClanID, 1)
		if err != nil {
			return nil, err
		}
	}
	return membership, nil
}

func recreateDeletedMembershipHelper(db DB, membership *Membership, level string, requestorID int, approved bool) (*Membership, error) {
	membership.RequestorID = requestorID
	membership.Level = level
	membership.Approved = approved
	membership.Denied = false

	_, err := db.Update(membership)
	if err != nil {
		return nil, err
	}
	if approved {
		err = IncrementPlayerMembershipCount(db, membership.PlayerID, 1)
		if err != nil {
			return nil, err
		}
		err = IncrementClanMembershipCount(db, membership.ClanID, 1)
		if err != nil {
			return nil, err
		}
	}
	return membership, nil
}

func promoteOrDemoteMemberHelper(db DB, membership *Membership, action string, levels util.JSON) (*Membership, error) {
	levelInt := GetLevelIntByLevel(membership.Level, levels)
	if action == "promote" {
		membership.Level = GetLevelByLevelInt(levelInt+1, levels)
	} else if action == "demote" {
		membership.Level = GetLevelByLevelInt(levelInt-1, levels)
	} else {
		return nil, &InvalidMembershipActionError{action}
	}

	if membership.Level == "" {
		return nil, &InvalidLevelForGameError{membership.GameID, membership.Level}
	}

	_, err := db.Update(membership)
	if err != nil {
		return nil, err
	}
	return membership, nil
}

func deleteMembershipHelper(db DB, membership *Membership, deletedBy int) error {
	if membership.Approved {
		err := IncrementPlayerMembershipCount(db, membership.PlayerID, -1)
		if err != nil {
			return err
		}
		err = IncrementClanMembershipCount(db, membership.ClanID, -1)
		if err != nil {
			return err
		}
	}

	membership.DeletedAt = time.Now().UnixNano() / 1000000
	membership.DeletedBy = deletedBy
	membership.Approved = false
	membership.Denied = false

	membership.Banned = deletedBy != membership.PlayerID // TODO: Test this

	_, err := db.Update(membership)
	return err
}

// GetLevelByLevelInt returns the level string given the level int
func GetLevelByLevelInt(levelInt int, levels util.JSON) string {
	for k, v := range levels {
		switch v.(type) {
		case float64:
			if int(v.(float64)) == levelInt {
				return k
			}
		case int:
			if v.(int) == levelInt {
				return k
			}
		}
	}
	return ""
}

// GetLevelIntByLevel returns the level string given the level int
func GetLevelIntByLevel(level string, levels util.JSON) int {
	v := levels[level]
	switch v.(type) {
	case float64:
		return int(v.(float64))
	default:
		return v.(int)
	}
}
