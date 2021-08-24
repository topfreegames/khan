// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"database/sql"

	"github.com/topfreegames/khan/util"

	"github.com/go-gorp/gorp"
)

var approveString = "approve"

// Membership relates a player to a clan
type Membership struct {
	ID          int64         `db:"id"`
	GameID      string        `db:"game_id"`
	Level       string        `db:"membership_level"`
	Approved    bool          `db:"approved"`
	Denied      bool          `db:"denied"`
	Banned      bool          `db:"banned"`
	PlayerID    int64         `db:"player_id"`
	ClanID      int64         `db:"clan_id"`
	RequestorID int64         `db:"requestor_id"`
	ApproverID  sql.NullInt64 `db:"approver_id"`
	DenierID    sql.NullInt64 `db:"denier_id"`
	CreatedAt   int64         `db:"created_at"`
	UpdatedAt   int64         `db:"updated_at"`
	DeletedBy   int64         `db:"deleted_by"`
	DeletedAt   int64         `db:"deleted_at"`
	ApprovedAt  int64         `db:"approved_at"`
	DeniedAt    int64         `db:"denied_at"`
	Message     string        `db:"message"`
}

// PreInsert populates fields before inserting a new clan
func (m *Membership) PreInsert(s gorp.SqlExecutor) error {
	if m.CreatedAt == 0 {
		m.CreatedAt = util.NowMilli()
	}
	m.UpdatedAt = m.CreatedAt
	return nil
}

// PreUpdate populates fields before updating a clan
func (m *Membership) PreUpdate(s gorp.SqlExecutor) error {
	m.UpdatedAt = util.NowMilli()
	return nil
}

// GetMembershipByID returns a membership by id
func GetMembershipByID(db DB, id int64) (*Membership, error) {
	obj, err := db.Get(Membership{}, id)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, &ModelNotFoundError{"Membership", id}
	}
	return obj.(*Membership), nil
}

// GetValidMembershipByClanAndPlayerPublicID returns a non deleted membership for the clan and the player with the given publicIDs
func GetValidMembershipByClanAndPlayerPublicID(db DB, gameID, clanPublicID, playerPublicID string) (*Membership, error) {
	var memberships []*Membership
	query := `
	SELECT
		m.*
	FROM memberships m
		INNER JOIN clans c ON c.game_id=$3 AND c.public_id=$1 AND c.id=m.clan_id
		INNER JOIN players p ON p.game_id=$3 AND p.public_id=$2 AND p.id=m.player_id
	WHERE
		m.game_id=$3 AND
		m.deleted_at=0`

	_, err := db.Select(&memberships, query, clanPublicID, playerPublicID, gameID)
	if err != nil {
		return nil, err
	}
	if memberships == nil || len(memberships) < 1 {
		return nil, &ModelNotFoundError{"Membership", playerPublicID}
	}
	return memberships[0], nil
}

// GetMembershipByClanAndPlayerPublicID returns a deleted membership for the clan and the player with the given publicIDs
func GetMembershipByClanAndPlayerPublicID(db DB, gameID, clanPublicID, playerPublicID string) (*Membership, error) {
	var memberships []*Membership
	query := `
	SELECT
		m.*
	FROM memberships m
		INNER JOIN clans c ON c.game_id=$3 AND c.public_id=$1 AND c.id=m.clan_id
		INNER JOIN players p ON p.game_id=$3 AND p.public_id=$2 AND p.id=m.player_id
	WHERE m.game_id=$3`

	_, err := db.Select(&memberships, query, clanPublicID, playerPublicID, gameID)
	if err != nil {
		return nil, err
	}
	if memberships == nil || len(memberships) < 1 {
		return nil, &ModelNotFoundError{"Membership", playerPublicID}
	}
	return memberships[0], nil
}

// GetDeletedMembershipByClanAndPlayerID returns a deleted membership for the player with the given ID and the clan ID
func GetDeletedMembershipByClanAndPlayerID(db DB, gameID string, clanID, playerID int64) (*Membership, error) {
	var memberships []*Membership
	query := `
	SELECT
		m.*
	FROM memberships m
		INNER JOIN clans c ON c.game_id=$3 AND c.id=$1 AND c.id=m.clan_id
		INNER JOIN players p ON p.game_id=$3 AND p.id=$2 AND p.id=m.player_id
	WHERE
		m.deleted_at!=0 AND m.game_id=$3`

	_, err := db.Select(&memberships, query, clanID, playerID, gameID)
	if err != nil {
		return nil, err
	}
	if memberships == nil || len(memberships) < 1 {
		return nil, &ModelNotFoundError{"Membership", playerID}
	}
	return memberships[0], nil
}

// GetOldestMemberWithHighestLevel returns the member with highest level that has the oldest creation date
func GetOldestMemberWithHighestLevel(db DB, gameID, clanPublicID string) (*Membership, error) {
	var memberships []*Membership
	query := `
	SELECT
	 m.*
	FROM memberships m
	 INNER JOIN games g ON g.public_id=m.game_id AND g.public_id=$1
	 INNER JOIN clans c ON c.game_id=$1 AND c.public_id=$2 AND c.id=m.clan_id
	WHERE m.deleted_at=0 AND m.approved=true
	ORDER BY
	 g.membership_levels::json->>m.membership_level DESC,
	 m.created_at ASC
	LIMIT 1`
	_, err := db.Select(&memberships, query, gameID, clanPublicID)
	if err != nil {
		return nil, err
	}
	if memberships == nil || len(memberships) < 1 {
		return nil, &ClanHasNoMembersError{clanPublicID}
	}
	return memberships[0], nil
}

// GetNumberOfPendingInvites gets total number of pending invites for player
func GetNumberOfPendingInvites(db DB, player *Player) (int, error) {
	membershipCount, err := db.SelectInt(`
		SELECT COUNT(*)
		FROM memberships m
		WHERE
			m.player_id = $1 AND m.player_id != m.requestor_id AND m.deleted_at = 0 AND
			m.approved = false AND m.denied = false AND m.banned = false
	`, player.ID)
	if err != nil {
		return -1, nil
	}
	return int(membershipCount), nil
}

// ApproveOrDenyMembershipInvitation sets Membership.Approved to true or Membership.Denied to true
func ApproveOrDenyMembershipInvitation(db DB, encryptionKey []byte, game *Game, gameID, playerPublicID, clanPublicID, action string) (*Membership, error) {
	membership, err := GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
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

	player, err := GetPlayerByID(db, encryptionKey, membership.PlayerID)
	if err != nil {
		return nil, err
	}
	if action == approveString {
		err = playerReachedMaxClans(db, encryptionKey, game, player)
		if err != nil {
			return nil, err
		}
		reachedMaxMembersError := clanReachedMaxMemberships(db, game, nil, membership.ClanID)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
	}
	return approveOrDenyMembershipHelper(db, membership, action, player)
}

// ApproveOrDenyMembershipApplication sets Membership.Approved to true or Membership.Denied to true
func ApproveOrDenyMembershipApplication(db DB, encryptionKey []byte, game *Game, gameID, playerPublicID, clanPublicID, requestorPublicID, action string) (*Membership, error) {
	if playerPublicID == requestorPublicID {
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
	}

	requestor, err := GetPlayerByPublicID(db, encryptionKey, gameID, requestorPublicID)
	if err != nil {
		return nil, err
	}

	membership, err := GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, err
	}

	if membership.PlayerID != membership.RequestorID {
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
	}

	if membership.Approved || membership.Denied {
		return nil, &CannotApproveOrDenyMembershipAlreadyProcessedError{action}
	}

	if action == approveString {
		player, err := GetPlayerByID(db, encryptionKey, membership.PlayerID)
		if err != nil {
			return nil, err
		}
		err = playerReachedMaxClans(db, encryptionKey, game, player)
		if err != nil {
			return nil, err
		}
		reachedMaxMembersError := clanReachedMaxMemberships(db, game, nil, membership.ClanID)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
	}

	reqMembership, _ := GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		_, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
		}
		return approveOrDenyMembershipHelper(db, membership, action, requestor)
	}

	levelInt := getLevelIntByLevel(reqMembership.Level, game.MembershipLevels)
	if !reqMembership.Approved || levelInt < game.MinLevelToAcceptApplication {
		return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
	}
	return approveOrDenyMembershipHelper(db, membership, action, requestor)
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

	membership, err := GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, err
	}
	if !isValidMember(membership) {
		return nil, &CannotPromoteOrDemoteInvalidMemberError{action}
	}

	levelInt := getLevelIntByLevel(membership.Level, game.MembershipLevels)
	if promote && levelInt >= game.MaxMembershipLevel || demote && levelInt <= game.MinMembershipLevel {
		return nil, &CannotPromoteOrDemoteMemberLevelError{action, levelInt}
	}

	reqMembership, _ := GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		_, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
		}
		return promoteOrDemoteMemberHelper(db, membership, action, game.MembershipLevels)
	}

	reqLevelInt := getLevelIntByLevel(reqMembership.Level, game.MembershipLevels)
	if isValidMember(reqMembership) && reqLevelInt >= levelInt+levelOffset {
		return promoteOrDemoteMemberHelper(db, membership, action, game.MembershipLevels)
	}
	return nil, &PlayerCannotPerformMembershipActionError{action, playerPublicID, clanPublicID, requestorPublicID}
}

// DeleteMembership soft deletes a membership
func DeleteMembership(db DB, game *Game, gameID, playerPublicID, clanPublicID, requestorPublicID string) (*Membership, error) {
	membership, err := GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, playerPublicID)
	if err != nil {
		return nil, err
	}
	if playerPublicID == requestorPublicID {
		return deleteMembershipHelper(db, membership, membership.PlayerID)
	}
	reqMembership, _ := GetValidMembershipByClanAndPlayerPublicID(db, gameID, clanPublicID, requestorPublicID)
	if reqMembership == nil {
		clan, clanErr := GetClanByPublicIDAndOwnerPublicID(db, gameID, clanPublicID, requestorPublicID)
		if clanErr != nil {
			return nil, &PlayerCannotPerformMembershipActionError{"delete", playerPublicID, clanPublicID, requestorPublicID}
		}
		return deleteMembershipHelper(db, membership, clan.OwnerID)
	}

	levelInt := getLevelIntByLevel(membership.Level, game.MembershipLevels)
	reqLevelInt := getLevelIntByLevel(reqMembership.Level, game.MembershipLevels)
	if isValidMember(reqMembership) && reqLevelInt >= game.MinLevelToRemoveMember && reqLevelInt >= levelInt+game.MinLevelOffsetToRemoveMember {
		return deleteMembershipHelper(db, membership, reqMembership.PlayerID)
	}
	return nil, &PlayerCannotPerformMembershipActionError{"delete", playerPublicID, clanPublicID, requestorPublicID}
}

// CreateMembership creates a new membership
func CreateMembership(db DB, encryptionKey []byte, game *Game, gameID, level, playerPublicID, clanPublicID, requestorPublicID, message string) (*Membership, error) {
	if _, levelValid := game.MembershipLevels[level]; !levelValid {
		return nil, &InvalidLevelForGameError{gameID, level}
	}

	clan, clanErr := GetClanByPublicID(db, game.PublicID, clanPublicID)
	if clanErr != nil {
		return nil, clanErr
	}

	membership, _ := GetMembershipByClanAndPlayerPublicID(db, gameID, clan.PublicID, playerPublicID)
	playerID, previousMembership, err := validateMembership(db, encryptionKey, game, membership, clan, playerPublicID, requestorPublicID)
	if err != nil {
		return nil, err
	}

	if clan.OwnerID == playerID {
		return nil, &AlreadyHasValidMembershipError{playerPublicID, clanPublicID}
	}

	if requestorPublicID == playerPublicID {
		return applyForMembership(db, game, membership, level, clan, playerID, requestorPublicID, message, previousMembership)
	}

	return inviteMember(db, encryptionKey, game, membership, level, clan, playerID, requestorPublicID, message, previousMembership)
}

func validateMembership(db DB, encryptionKey []byte, game *Game, membership *Membership, clan *Clan, playerPublicID, requestorPublicID string) (int64, bool, error) {
	playerID := int64(-1)
	previousMembership := false
	if membership != nil {
		previousMembership = true
		nowInMilliseconds := util.NowMilli()
		applicationInOpenClan := requestorPublicID == playerPublicID && clan.AllowApplication && clan.AutoJoin
		if membership.Approved {
			return -1, false, &AlreadyHasValidMembershipError{playerPublicID, clan.PublicID}
		} else if !applicationInOpenClan && membership.Denied && membership.DenierID.Int64 != membership.PlayerID {
			timeToBeReady := game.CooldownAfterDeny - int(nowInMilliseconds-membership.DeniedAt)/1000
			if timeToBeReady > 0 {
				return -1, false, &MustWaitMembershipCooldownError{timeToBeReady, playerPublicID, clan.PublicID}
			}
		} else if membership.DeletedAt > 0 && membership.DeletedBy != membership.PlayerID && playerPublicID == requestorPublicID {
			// Allow immediate membership creation if player is being invited
			timeToBeReady := game.CooldownAfterDelete - int(nowInMilliseconds-membership.DeletedAt)/1000
			if timeToBeReady > 0 {
				return -1, false, &MustWaitMembershipCooldownError{timeToBeReady, playerPublicID, clan.PublicID}
			}
		} else {
			// TODO: When allowing 'memberLeft' players to apply we do not avoid flooding in this case =/
			memberLeft := membership.DeletedAt > 0 && membership.DeletedBy == membership.PlayerID

			cd := 0
			previousInvite := membership.RequestorID != membership.PlayerID

			// invite and previous invite
			if previousInvite && requestorPublicID != playerPublicID {
				cd = game.CooldownBeforeInvite
			}
			// application and previous application
			if !previousInvite && requestorPublicID == playerPublicID && !memberLeft {
				cd = game.CooldownBeforeApply
			}

			if cd != 0 {
				timeToBeReady := cd - int(nowInMilliseconds-membership.UpdatedAt)/1000
				if timeToBeReady > 0 {
					return -1, false, &MustWaitMembershipCooldownError{timeToBeReady, playerPublicID, clan.PublicID}
				}
			}
		}

		playerID = membership.PlayerID
		player, err := GetPlayerByID(db, encryptionKey, playerID)
		if err != nil {
			return -1, false, err
		}
		err = playerReachedMaxClans(db, encryptionKey, game, player)
		if err != nil {
			return -1, false, err
		}
	} else {
		player, err := GetPlayerByPublicID(db, encryptionKey, game.PublicID, playerPublicID)
		if err != nil {
			return -1, false, err
		}
		playerID = player.ID
		err = playerReachedMaxClans(db, encryptionKey, game, player)
		if err != nil {
			return -1, false, err
		}
	}

	return playerID, previousMembership, nil
}

func applyForMembership(db DB, game *Game, membership *Membership, level string, clan *Clan, playerID int64, requestorPublicID, message string, previousMembership bool) (*Membership, error) {
	if !clan.AllowApplication {
		return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clan.PublicID}
	}

	reachedMaxMembersError := clanReachedMaxMemberships(db, game, clan, -1)
	if reachedMaxMembersError != nil {
		return nil, reachedMaxMembersError
	}
	if previousMembership {
		return updatePreviousMembershipHelper(db, membership, level, membership.PlayerID, message, clan.AutoJoin)
	}
	return createMembershipHelper(db, game.PublicID, level, playerID, clan.ID, playerID, message, clan.AutoJoin)
}

func inviteMember(db DB, encryptionKey []byte, game *Game, membership *Membership, level string, clan *Clan, playerID int64, requestorPublicID, message string, previousMembership bool) (*Membership, error) {
	reqMembership, _ := GetValidMembershipByClanAndPlayerPublicID(db, game.PublicID, clan.PublicID, requestorPublicID)
	if reqMembership == nil {
		requestor, err := GetPlayerByPublicID(db, encryptionKey, game.PublicID, requestorPublicID)
		if err != nil {
			return nil, err
		}
		// Did not find a memebership and player is not clan owner
		if requestor.ID != clan.OwnerID {
			return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clan.PublicID}
		}
		reachedMaxInvitesError := playerReachedMaxInvites(db, encryptionKey, game, playerID)
		if reachedMaxInvitesError != nil {
			return nil, reachedMaxInvitesError
		}
		reachedMaxMembersError := clanReachedMaxMemberships(db, game, clan, -1)
		if reachedMaxMembersError != nil {
			return nil, reachedMaxMembersError
		}
		if previousMembership {
			return updatePreviousMembershipHelper(db, membership, level, clan.OwnerID, message, false)
		}
		return createMembershipHelper(db, game.PublicID, level, playerID, clan.ID, clan.OwnerID, message, false)
	}

	reachedMaxMembersError := clanReachedMaxMemberships(db, game, nil, reqMembership.ClanID)
	if reachedMaxMembersError != nil {
		return nil, reachedMaxMembersError
	}

	levelInt := getLevelIntByLevel(reqMembership.Level, game.MembershipLevels)

	if isValidMember(reqMembership) && levelInt >= game.MinLevelToCreateInvitation {
		if previousMembership {
			return updatePreviousMembershipHelper(db, membership, level, reqMembership.PlayerID, message, false)
		}
		return createMembershipHelper(db, game.PublicID, level, playerID, reqMembership.ClanID, reqMembership.PlayerID, message, false)
	}
	return nil, &PlayerCannotCreateMembershipError{requestorPublicID, clan.PublicID}
}

func clanReachedMaxMemberships(db DB, game *Game, clan *Clan, clanID int64) error {
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

func playerReachedMaxInvites(db DB, encryptionKey []byte, game *Game, playerID int64) error {
	player, err := GetPlayerByID(db, encryptionKey, playerID)
	if err != nil {
		return err
	}

	pendingInvites, err := GetNumberOfPendingInvites(db, player)
	if err != nil {
		return err
	}

	if game.MaxPendingInvites > 0 && pendingInvites >= game.MaxPendingInvites {
		return &PlayerReachedMaxInvitesError{ID: player.PublicID}
	}

	return nil
}

func playerReachedMaxClans(db DB, encryptionKey []byte, game *Game, player *Player) error {
	playerID := player.ID
	if player.MembershipCount+player.OwnershipCount >= game.MaxClansPerPlayer {
		err := UpdatePlayerMembershipCount(db, playerID)
		if err != nil {
			return err
		}
		player, err = GetPlayerByID(db, encryptionKey, playerID)
		if err != nil {
			return err
		}
		if player.MembershipCount+player.OwnershipCount >= game.MaxClansPerPlayer {
			return &PlayerReachedMaxClansError{player.PublicID}
		}
	}
	return nil
}

func isValidMember(membership *Membership) bool {
	return membership.Approved && !membership.Denied
}

func approveOrDenyMembershipHelper(db DB, membership *Membership, action string, performer *Player) (*Membership, error) {
	approve := action == approveString
	if approve {
		membership.Approved = true
		membership.ApproverID = sql.NullInt64{Int64: performer.ID, Valid: true}
		membership.ApprovedAt = util.NowMilli()
	} else if action == "deny" {
		membership.Denied = true
		membership.DenierID = sql.NullInt64{Int64: performer.ID, Valid: true}
		membership.DeniedAt = util.NowMilli()

	} else {
		return nil, &InvalidMembershipActionError{action}
	}
	_, err := db.Update(membership)
	if err != nil {
		return nil, err
	}
	if approve {
		err = UpdatePlayerMembershipCount(db, membership.PlayerID)
		if err != nil {
			return nil, err
		}
		err = UpdateClanMembershipCount(db, membership.ClanID)
		if err != nil {
			return nil, err
		}
	}
	return membership, nil
}

func createMembershipHelper(db DB, gameID, level string, playerID, clanID, requestorID int64, message string, approved bool) (*Membership, error) {
	membership := &Membership{
		GameID:      gameID,
		ClanID:      clanID,
		PlayerID:    playerID,
		RequestorID: requestorID,
		Level:       level,
		Approved:    approved,
		Denied:      false,
		Message:     message,
	}

	if approved {
		membership.ApproverID = sql.NullInt64{Int64: requestorID, Valid: true}
		membership.ApprovedAt = util.NowMilli()
	}
	err := db.Insert(membership)
	if err != nil {
		return nil, err
	}
	if approved {
		err = UpdatePlayerMembershipCount(db, membership.PlayerID)
		if err != nil {
			return nil, err
		}
		err = UpdateClanMembershipCount(db, membership.ClanID)
		if err != nil {
			return nil, err
		}
	}
	return membership, nil
}

func updatePreviousMembershipHelper(db DB, membership *Membership, level string, requestorID int64, message string, approved bool) (*Membership, error) {
	membership.RequestorID = requestorID
	membership.Level = level
	membership.Approved = approved
	membership.Denied = false
	membership.Banned = false
	membership.DeletedAt = 0
	membership.DeletedBy = 0
	membership.Message = message
	if approved {
		membership.ApproverID = sql.NullInt64{Int64: requestorID, Valid: true}
		membership.ApprovedAt = util.NowMilli()
	}

	_, err := db.Update(membership)
	if err != nil {
		return nil, err
	}
	if approved {
		err = UpdatePlayerMembershipCount(db, membership.PlayerID)
		if err != nil {
			return nil, err
		}
		err = UpdateClanMembershipCount(db, membership.ClanID)
		if err != nil {
			return nil, err
		}
	}
	return membership, nil
}

func promoteOrDemoteMemberHelper(db DB, membership *Membership, action string, levels map[string]interface{}) (*Membership, error) {
	levelInt := getLevelIntByLevel(membership.Level, levels)
	if action == "promote" {
		membership.Level = getLevelByLevelInt(levelInt+1, levels)
	} else if action == "demote" {
		membership.Level = getLevelByLevelInt(levelInt-1, levels)
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

func deleteMembershipHelper(db DB, membership *Membership, deletedBy int64) (*Membership, error) {
	membershipWasApproved := membership.Approved
	membership.DeletedAt = util.NowMilli()
	membership.DeletedBy = deletedBy
	membership.Approved = false
	membership.Denied = false

	membership.Banned = deletedBy != membership.PlayerID // TODO: Test this

	_, err := db.Update(membership)
	if err != nil {
		return nil, err
	}

	if membershipWasApproved {
		err := UpdatePlayerMembershipCount(db, membership.PlayerID)
		if err != nil {
			return nil, err
		}
		err = UpdateClanMembershipCount(db, membership.ClanID)
		if err != nil {
			return nil, err
		}
	}
	return membership, err
}

func getLevelByLevelInt(levelInt int, levels map[string]interface{}) string {
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

func getLevelIntByLevel(level string, levels map[string]interface{}) int {
	v := levels[level]
	switch v.(type) {
	case float64:
		return int(v.(float64))
	default:
		return v.(int)
	}
}
