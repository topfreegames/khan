// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"encoding/json"
	"fmt"

	"github.com/topfreegames/khan/util"

	"github.com/go-gorp/gorp"
)

// Game identifies uniquely one game
type Game struct {
	ID                                             int                    `db:"id"`
	PublicID                                       string                 `db:"public_id"`
	Name                                           string                 `db:"name"`
	MinMembershipLevel                             int                    `db:"min_membership_level"`
	MaxMembershipLevel                             int                    `db:"max_membership_level"`
	MinLevelToAcceptApplication                    int                    `db:"min_level_to_accept_application"`
	MinLevelToCreateInvitation                     int                    `db:"min_level_to_create_invitation"`
	MinLevelToRemoveMember                         int                    `db:"min_level_to_remove_member"`
	MinLevelOffsetToRemoveMember                   int                    `db:"min_level_offset_to_remove_member"`
	MinLevelOffsetToPromoteMember                  int                    `db:"min_level_offset_to_promote_member"`
	MinLevelOffsetToDemoteMember                   int                    `db:"min_level_offset_to_demote_member"`
	MaxMembers                                     int                    `db:"max_members"`
	MaxClansPerPlayer                              int                    `db:"max_clans_per_player"`
	MembershipLevels                               map[string]interface{} `db:"membership_levels"`
	Metadata                                       map[string]interface{} `db:"metadata"`
	CreatedAt                                      int64                  `db:"created_at"`
	UpdatedAt                                      int64                  `db:"updated_at"`
	CooldownAfterDeny                              int                    `db:"cooldown_after_deny"`
	CooldownAfterDelete                            int                    `db:"cooldown_after_delete"`
	CooldownBeforeApply                            int                    `db:"cooldown_before_apply"`
	CooldownBeforeInvite                           int                    `db:"cooldown_before_invite"`
	MaxPendingInvites                              int                    `db:"max_pending_invites"`
	ClanUpdateMetadataFieldsHookTriggerWhitelist   string                 `db:"clan_metadata_fields_whitelist"`
	PlayerUpdateMetadataFieldsHookTriggerWhitelist string                 `db:"player_metadata_fields_whitelist"`
}

// PreInsert populates fields before inserting a new game
func (g *Game) PreInsert(s gorp.SqlExecutor) error {
	// Handle JSON fields
	sortedLevels := util.SortLevels(g.MembershipLevels)
	g.MinMembershipLevel = sortedLevels[0].Value
	g.MaxMembershipLevel = sortedLevels[len(sortedLevels)-1].Value
	g.CreatedAt = util.NowMilli()
	g.UpdatedAt = g.CreatedAt
	return nil
}

// PreUpdate populates fields before updating a game
func (g *Game) PreUpdate(s gorp.SqlExecutor) error {
	sortedLevels := util.SortLevels(g.MembershipLevels)
	g.MinMembershipLevel = sortedLevels[0].Value
	g.MaxMembershipLevel = sortedLevels[len(sortedLevels)-1].Value
	g.UpdatedAt = util.NowMilli()
	return nil
}

// GetGameByID returns a game by id
func GetGameByID(db DB, id int) (*Game, error) {
	obj, err := db.Get(Game{}, id)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, &ModelNotFoundError{"Game", id}
	}

	game := obj.(*Game)
	return game, nil
}

// GetGameByPublicID returns a game by their public id
func GetGameByPublicID(db DB, publicID string) (*Game, error) {
	var games []*Game
	_, err := db.Select(&games, "SELECT * FROM games WHERE public_id=$1", publicID)
	if err != nil {
		return nil, err
	}
	if games == nil || len(games) < 1 {
		return nil, &ModelNotFoundError{"Game", publicID}
	}
	return games[0], nil
}

// GetAllGames returns all games in the DB
func GetAllGames(db DB) ([]*Game, error) {
	var games []*Game
	_, err := db.Select(&games, "SELECT * FROM games")
	if err != nil {
		return nil, err
	}
	return games, nil
}

// CreateGame creates a new game
func CreateGame(
	db DB,
	publicID, name string,
	levels, metadata map[string]interface{},
	minLevelAccept, minLevelCreate, minLevelRemove,
	minOffsetRemove, minOffsetPromote, minOffsetDemote, maxMembers,
	maxClans, cooldownAfterDeny, cooldownAfterDelete, cooldownBeforeApply,
	cooldownBeforeInvite, maxPendingInvites int, upsert bool,
	clanUpdateMetadataFieldsHookTriggerWhitelist string,
	playerUpdateMetadataFieldsHookTriggerWhitelist string,
) (*Game, error) {
	levelsJSON, err := json.Marshal(levels)
	if err != nil {
		return nil, err
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	sortedLevels := util.SortLevels(levels)
	minMembershipLevel := sortedLevels[0].Value
	maxMembershipLevel := sortedLevels[len(sortedLevels)-1].Value

	query := `
			INSERT INTO games(
				public_id,
				name,
				min_level_to_accept_application,
				min_level_to_create_invitation,
				min_level_to_remove_member,
				min_level_offset_to_remove_member,
				min_level_offset_to_promote_member,
				min_level_offset_to_demote_member,
				max_members,
				max_clans_per_player,
				membership_levels,
				metadata,
				cooldown_after_delete,
				cooldown_after_deny,
				cooldown_before_apply,
				cooldown_before_invite,
				min_membership_level,
				max_membership_level,
				max_pending_invites,
				clan_metadata_fields_whitelist,
				player_metadata_fields_whitelist,
				created_at,
				updated_at
			)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $22)%s`
	onConflict := ` ON CONFLICT (public_id)
			DO UPDATE set
				name=$2,
				min_level_to_accept_application=$3,
				min_level_to_create_invitation=$4,
				min_level_to_remove_member=$5,
				min_level_offset_to_remove_member=$6,
				min_level_offset_to_promote_member=$7,
				min_level_offset_to_demote_member=$8,
				max_members=$9,
				max_clans_per_player=$10,
				membership_levels=$11,
				metadata=$12,
				cooldown_after_delete=$13,
				cooldown_after_deny=$14,
				cooldown_before_apply=$15,
				cooldown_before_invite=$16,
				min_membership_level=$17,
				max_membership_level=$18,
				max_pending_invites=$19,
				clan_metadata_fields_whitelist=$20,
				player_metadata_fields_whitelist=$21,
				updated_at=$22
			WHERE games.public_id=$1`

	if upsert {
		query = fmt.Sprintf(query, onConflict)
	} else {
		query = fmt.Sprintf(query, "")
	}

	_, err = db.Exec(query,
		publicID,             // $1
		name,                 // $2
		minLevelAccept,       // $3
		minLevelCreate,       // $4
		minLevelRemove,       // $5
		minOffsetRemove,      // $6
		minOffsetPromote,     // $7
		minOffsetDemote,      // $8
		maxMembers,           // $9
		maxClans,             // $10
		levelsJSON,           // $11
		metadataJSON,         // $12
		cooldownAfterDelete,  // $13
		cooldownAfterDeny,    // $14
		cooldownBeforeApply,  // $15
		cooldownBeforeInvite, // $16
		minMembershipLevel,   // $17
		maxMembershipLevel,   // $18
		maxPendingInvites,    // $19
		clanUpdateMetadataFieldsHookTriggerWhitelist,   // $20
		playerUpdateMetadataFieldsHookTriggerWhitelist, // $21
		util.NowMilli(), // $22
	)
	if err != nil {
		return nil, err
	}
	return GetGameByPublicID(db, publicID)
}

// UpdateGame updates an existing game
func UpdateGame(
	db DB, publicID, name string, levels, metadata map[string]interface{},
	minLevelAccept, minLevelCreate, minLevelRemove, minOffsetRemove, minOffsetPromote,
	minOffsetDemote, maxMembers, maxClans, cooldownAfterDeny, cooldownAfterDelete,
	cooldownBeforeApply, cooldownBeforeInvite, maxPendingInvites int,
	clanUpdateMetadataFieldsHookTriggerWhitelist string,
	playerUpdateMetadataFieldsHookTriggerWhitelist string,
) (*Game, error) {
	return CreateGame(
		db, publicID, name, levels, metadata, minLevelAccept, minLevelCreate,
		minLevelRemove, minOffsetRemove, minOffsetPromote, minOffsetDemote,
		maxMembers, maxClans, cooldownAfterDeny, cooldownAfterDelete, cooldownBeforeApply,
		cooldownBeforeInvite, maxPendingInvites, true,
		clanUpdateMetadataFieldsHookTriggerWhitelist,
		playerUpdateMetadataFieldsHookTriggerWhitelist,
	)
}
