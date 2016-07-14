// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"fmt"

	"github.com/topfreegames/khan/util"

	"gopkg.in/gorp.v1"
)

// Game identifies uniquely one game
type Game struct {
	ID                            int                    `db:"id"`
	PublicID                      string                 `db:"public_id"`
	Name                          string                 `db:"name"`
	MinMembershipLevel            int                    `db:"min_membership_level"`
	MaxMembershipLevel            int                    `db:"max_membership_level"`
	MinLevelToAcceptApplication   int                    `db:"min_level_to_accept_application"`
	MinLevelToCreateInvitation    int                    `db:"min_level_to_create_invitation"`
	MinLevelToRemoveMember        int                    `db:"min_level_to_remove_member"`
	MinLevelOffsetToRemoveMember  int                    `db:"min_level_offset_to_remove_member"`
	MinLevelOffsetToPromoteMember int                    `db:"min_level_offset_to_promote_member"`
	MinLevelOffsetToDemoteMember  int                    `db:"min_level_offset_to_demote_member"`
	MaxMembers                    int                    `db:"max_members"`
	MaxClansPerPlayer             int                    `db:"max_clans_per_player"`
	MembershipLevels              map[string]interface{} `db:"membership_levels"`
	Metadata                      map[string]interface{} `db:"metadata"`
	CreatedAt                     int64                  `db:"created_at"`
	UpdatedAt                     int64                  `db:"updated_at"`
	CooldownAfterDeny             int                    `db:"cooldown_after_deny"`
	CooldownAfterDelete           int                    `db:"cooldown_after_delete"`
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
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"Game", id}
	}

	game := obj.(*Game)
	return game, nil
}

// GetGameByPublicID returns a game by their public id
func GetGameByPublicID(db DB, publicID string) (*Game, error) {
	var game Game
	err := db.SelectOne(&game, "SELECT * FROM games WHERE public_id=$1", publicID)
	if err != nil || &game == nil {
		return nil, &ModelNotFoundError{"Game", publicID}
	}
	return &game, nil
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
func CreateGame(db DB, publicID, name string, levels, metadata map[string]interface{},
	minLevelAccept, minLevelCreate, minLevelRemove, minOffsetRemove, minOffsetPromote, minOffsetDemote, maxMembers, maxClans, cooldownAfterDeny, cooldownAfterDelete int,
) (*Game, error) {
	game := &Game{
		PublicID: publicID,
		Name:     name,
		MinLevelToAcceptApplication:   minLevelAccept,
		MinLevelToCreateInvitation:    minLevelCreate,
		MinLevelToRemoveMember:        minLevelRemove,
		MinLevelOffsetToRemoveMember:  minOffsetRemove,
		MinLevelOffsetToPromoteMember: minOffsetPromote,
		MinLevelOffsetToDemoteMember:  minOffsetDemote,
		MaxMembers:                    maxMembers,
		MaxClansPerPlayer:             maxClans,
		MembershipLevels:              levels,
		Metadata:                      metadata,
		CooldownAfterDelete:           cooldownAfterDelete,
		CooldownAfterDeny:             cooldownAfterDeny,
	}
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}
	return game, nil
}

// UpdateGame updates an existing game
func UpdateGame(db DB, publicID, name string, levels, metadata map[string]interface{},
	minLevelAccept, minLevelCreate, minLevelRemove, minOffsetRemove, minOffsetPromote, minOffsetDemote, maxMembers, maxClans, cooldownAfterDeny, cooldownAfterDelete int,
) (*Game, error) {
	game, err := GetGameByPublicID(db, publicID)

	if err != nil {
		if err.Error() == fmt.Sprintf("Game was not found with id: %s", publicID) {
			return CreateGame(
				db, publicID, name, levels, metadata, minLevelAccept,
				minLevelCreate, minLevelRemove, minOffsetRemove, minOffsetPromote,
				minOffsetDemote, maxMembers, maxClans, cooldownAfterDeny, cooldownAfterDelete,
			)
		}
		return nil, err
	}

	game.Name = name
	game.MinLevelToAcceptApplication = minLevelAccept
	game.MinLevelToCreateInvitation = minLevelCreate
	game.MinLevelToRemoveMember = minLevelRemove
	game.MinLevelOffsetToRemoveMember = minOffsetRemove
	game.MinLevelOffsetToPromoteMember = minOffsetPromote
	game.MinLevelOffsetToDemoteMember = minOffsetDemote
	game.MaxMembers = maxMembers
	game.MaxClansPerPlayer = maxClans
	game.MembershipLevels = levels
	game.Metadata = metadata
	game.CooldownAfterDeny = cooldownAfterDeny
	game.CooldownAfterDelete = cooldownAfterDelete

	_, err = db.Update(game)

	if err != nil {
		return nil, err
	}

	return game, nil
}
