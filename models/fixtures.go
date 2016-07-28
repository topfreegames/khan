// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import (
	"database/sql"
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/util"
)

// GameFactory is responsible for constructing test game instances
var GameFactory = factory.NewFactory(
	&Game{
		MinLevelToAcceptApplication:   2,
		MinLevelToCreateInvitation:    2,
		MinLevelToRemoveMember:        2,
		MinLevelOffsetToRemoveMember:  1,
		MinLevelOffsetToPromoteMember: 2,
		MinLevelOffsetToDemoteMember:  1,
		MaxClansPerPlayer:             1,
		MaxMembers:                    100,
		CooldownAfterDeny:             0,
		CooldownAfterDelete:           0,
	},
).Attr("PublicID", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("Name", func(args factory.Args) (interface{}, error) {
	return uuid.NewV4().String(), nil
}).Attr("Metadata", func(args factory.Args) (interface{}, error) {
	return map[string]interface{}{}, nil
}).Attr("MembershipLevels", func(args factory.Args) (interface{}, error) {
	return map[string]interface{}{"Member": 1, "Elder": 2, "CoLeader": 3}, nil
})

// HookFactory is responsible for constructing event hook instances
var HookFactory = factory.NewFactory(
	&Hook{EventType: GameUpdatedHook, URL: "http://test/game-created"},
)

// CreateHookFactory is responsible for creating a test hook instance with the associated game
func CreateHookFactory(db DB, gameID string, eventType int, url string) (*Hook, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	hook := HookFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":    gameID,
		"PublicID":  uuid.NewV4().String(),
		"EventType": eventType,
		"URL":       url,
	}).(*Hook)
	err = db.Insert(hook)
	if err != nil {
		return nil, err
	}
	return hook, nil
}

// GetHooksForRoutes gets hooks for all the specified routes
func GetHooksForRoutes(db DB, routes []string, eventType int) ([]*Hook, error) {
	var hooks []*Hook

	gameID := uuid.NewV4().String()

	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	for _, route := range routes {
		hook := HookFactory.MustCreateWithOption(map[string]interface{}{
			"GameID":    gameID,
			"PublicID":  uuid.NewV4().String(),
			"EventType": eventType,
			"URL":       route,
		}).(*Hook)
		err := db.Insert(hook)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}

	return hooks, nil
}

func configureFactory(fct *factory.Factory) *factory.Factory {
	return fct.Attr("PublicID", func(args factory.Args) (interface{}, error) {
		return uuid.NewV4().String(), nil
	}).Attr("Name", func(args factory.Args) (interface{}, error) {
		return randomdata.FullName(randomdata.RandomGender), nil
	}).Attr("Metadata", func(args factory.Args) (interface{}, error) {
		return map[string]interface{}{}, nil
	})
}

// PlayerFactory is responsible for constructing test player instances
var PlayerFactory = configureFactory(factory.NewFactory(
	&Player{},
))

// ClanFactory is responsible for constructing test clan instances
var ClanFactory = configureFactory(factory.NewFactory(
	&Clan{},
))

// CreatePlayerFactory is responsible for creating a test player instance with the associated game
func CreatePlayerFactory(db DB, gameID string, skipCreateGame ...bool) (*Game, *Player, error) {
	var game *Game

	if skipCreateGame == nil || len(skipCreateGame) != 1 || !skipCreateGame[0] {
		if gameID == "" {
			gameID = uuid.NewV4().String()
		}
		game = GameFactory.MustCreateWithOption(map[string]interface{}{
			"PublicID": gameID,
		}).(*Game)
		err := db.Insert(game)
		if err != nil {
			return nil, nil, err
		}
	} else {
		var err error
		game, err = GetGameByPublicID(db, gameID)
		if err != nil {
			return nil, nil, err
		}
	}

	player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID": gameID,
	}).(*Player)
	err := db.Insert(player)
	if err != nil {
		return nil, nil, err
	}
	return game, player, nil
}

// MembershipFactory is responsible for constructing test membership instances
var MembershipFactory = factory.NewFactory(
	&Membership{},
).SeqInt("GameID", func(n int) (interface{}, error) {
	return fmt.Sprintf("game-%d", n), nil
}).Attr("ApproverID", func(args factory.Args) (interface{}, error) {
	membership := args.Instance().(*Membership)
	approverID := 0
	valid := false
	if membership.Approved {
		approverID = membership.RequestorID
		valid = true
	}
	return sql.NullInt64{Int64: int64(approverID), Valid: valid}, nil
}).Attr("DenierID", func(args factory.Args) (interface{}, error) {
	membership := args.Instance().(*Membership)
	denierID := 0
	valid := false
	if membership.Denied {
		denierID = membership.RequestorID
		valid = true
	}
	return sql.NullInt64{Int64: int64(denierID), Valid: valid}, nil
})

// GetClanWithMemberships returns a clan filled with the number of memberships specified
func GetClanWithMemberships(
	db DB, approvedMemberships, deniedMemberships, bannedMemberships, pendingMemberships int, gameID string, clanPublicID string, options ...bool) (*Game, *Clan, *Player, []*Player, []*Membership, error) {
	var game *Game

	pendingsAreInvites := true
	if options != nil && len(options) > 1 {
		pendingsAreInvites = options[1]
	}

	if options == nil || len(options) == 0 || !options[0] {
		if gameID == "" {
			gameID = uuid.NewV4().String()
		}
		game = GameFactory.MustCreateWithOption(map[string]interface{}{
			"PublicID": gameID,
		}).(*Game)
		err := db.Insert(game)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	} else {
		var err error
		game, err = GetGameByPublicID(db, gameID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}
	if clanPublicID == "" {
		clanPublicID = uuid.NewV4().String()
	}

	owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":         gameID,
		"OwnershipCount": 1,
	}).(*Player)
	err := db.Insert(owner)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":          owner.GameID,
		"PublicID":        clanPublicID,
		"OwnerID":         owner.ID,
		"Metadata":        map[string]interface{}{"x": "a"},
		"MembershipCount": approvedMemberships + 1,
	}).(*Clan)
	err = db.Insert(clan)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	testData := []map[string]interface{}{
		map[string]interface{}{
			"approved": true,
			"denied":   false,
			"banned":   false,
			"count":    approvedMemberships,
		},
		map[string]interface{}{
			"approved": false,
			"denied":   true,
			"banned":   false,
			"count":    deniedMemberships,
		},
		map[string]interface{}{
			"approved": false,
			"denied":   false,
			"banned":   true,
			"count":    bannedMemberships,
		},
		map[string]interface{}{
			"approved": false,
			"denied":   false,
			"banned":   false,
			"count":    pendingMemberships,
		},
	}

	var players []*Player
	var memberships []*Membership

	for _, data := range testData {
		approved := data["approved"].(bool)
		denied := data["denied"].(bool)
		banned := data["banned"].(bool)
		count := data["count"].(int)

		for i := 0; i < count; i++ {
			membershipCount := 0
			if approved {
				membershipCount = 1
			}
			player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":          owner.GameID,
				"MembershipCount": membershipCount,
			}).(*Player)
			err = db.Insert(player)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			players = append(players, player)

			requestorID := owner.ID
			message := ""
			if !pendingsAreInvites {
				requestorID = player.ID
				message = "Accept me"
			}

			membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":      owner.GameID,
				"PlayerID":    player.ID,
				"ClanID":      clan.ID,
				"RequestorID": requestorID,
				"Metadata":    map[string]interface{}{"x": "a"},
				"Level":       "Member",
				"Approved":    approved,
				"Denied":      denied,
				"Banned":      banned,
				"Message":     message,
			}).(*Membership)

			if approved {
				approverID := player.ID
				if !pendingsAreInvites {
					approverID = owner.ID
				}

				membership.ApproverID = sql.NullInt64{Int64: int64(approverID), Valid: true}
				membership.ApprovedAt = util.NowMilli()
			} else if denied {
				denierID := player.ID
				if !pendingsAreInvites {
					denierID = owner.ID
				}

				membership.DenierID = sql.NullInt64{Int64: int64(denierID), Valid: true}
				membership.DeniedAt = util.NowMilli()
			}

			err = db.Insert(membership)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}

			memberships = append(memberships, membership)
		}
	}

	return game, clan, owner, players, memberships, nil
}

// GetClanReachedMaxMemberships returns a clan with one approved membership, one unapproved membership and game MaxMembers=1
func GetClanReachedMaxMemberships(db DB) (*Game, *Clan, *Player, []*Player, []*Membership, error) {
	gameID := uuid.NewV4().String()
	clanPublicID := uuid.NewV4().String()

	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID":   gameID,
		"MaxMembers": 1,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":         gameID,
		"OwnershipCount": 1,
	}).(*Player)
	err = db.Insert(owner)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var players []*Player

	for i := 0; i < 2; i++ {
		membershipCount := 0
		if i == 0 {
			membershipCount = 1
		}
		player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
			"GameID":          owner.GameID,
			"MembershipCount": membershipCount,
		}).(*Player)
		err = db.Insert(player)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		players = append(players, player)
	}

	clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":          owner.GameID,
		"PublicID":        clanPublicID,
		"OwnerID":         owner.ID,
		"Metadata":        map[string]interface{}{"x": "a"},
		"MembershipCount": 2,
	}).(*Clan)
	err = db.Insert(clan)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var memberships []*Membership

	membership := MembershipFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":      owner.GameID,
		"PlayerID":    players[0].ID,
		"ClanID":      clan.ID,
		"RequestorID": owner.ID,
		"Metadata":    map[string]interface{}{"x": "a"},
		"Approved":    true,
		"Level":       "Member",
		"ApproverID":  sql.NullInt64{Int64: int64(players[0].ID), Valid: true},
		"ApprovedAt":  util.NowMilli(),
	}).(*Membership)
	err = db.Insert(membership)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	memberships = append(memberships, membership)

	membership = MembershipFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":      owner.GameID,
		"PlayerID":    players[1].ID,
		"ClanID":      clan.ID,
		"RequestorID": owner.ID,
		"Metadata":    map[string]interface{}{"x": "a"},
		"Approved":    false,
		"Level":       "Member",
	}).(*Membership)
	err = db.Insert(membership)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	memberships = append(memberships, membership)

	return game, clan, owner, players, memberships, nil
}

// GetTestClans returns a list of clans for tests
func GetTestClans(db DB, gameID string, publicIDTemplate string, numberOfClans int) (*Player, []*Clan, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, err
	}

	player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":         gameID,
		"OwnershipCount": numberOfClans,
	}).(*Player)
	err = db.Insert(player)
	if err != nil {
		return nil, nil, err
	}

	if publicIDTemplate == "" {
		publicIDTemplate = uuid.NewV4().String()
	}

	var clans []*Clan
	for i := 0; i < numberOfClans; i++ {
		clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
			"GameID":   player.GameID,
			"PublicID": fmt.Sprintf("%s-%d", publicIDTemplate, i),
			"Name":     fmt.Sprintf("💩clán-%s-%d", publicIDTemplate, i),
			"OwnerID":  player.ID,
		}).(*Clan)
		err = db.Insert(clan)
		if err != nil {
			return nil, nil, err
		}

		clans = append(clans, clan)
	}

	return player, clans, nil
}

// GetTestHooks return a fixed number of hooks for each event available
func GetTestHooks(db DB, gameID string, numberOfHooks int) ([]*Hook, error) {
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	var hooks []*Hook
	for i := 0; i < 2; i++ {
		for j := 0; j < numberOfHooks; j++ {
			hook := HookFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":    gameID,
				"PublicID":  uuid.NewV4().String(),
				"EventType": i,
				"URL":       fmt.Sprintf("http://test/event-%d-%d", i, j),
			}).(*Hook)
			err = db.Insert(hook)
			if err != nil {
				return nil, err
			}
			hooks = append(hooks, hook)
		}
	}

	return hooks, nil
}

//GetTestPlayerWithMemberships returns a player with approved, rejected and banned memberships
func GetTestPlayerWithMemberships(db DB, gameID string, approvedMemberships, rejectedMemberships, bannedMemberships, pendingMemberships int) (*Player, error) {
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":         gameID,
		"OwnershipCount": 1,
	}).(*Player)
	err = db.Insert(owner)
	if err != nil {
		return nil, err
	}

	player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":          owner.GameID,
		"MembershipCount": approvedMemberships,
	}).(*Player)
	err = db.Insert(player)
	if err != nil {
		return nil, err
	}

	createClan := func() (*Clan, error) {
		clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
			"GameID":          owner.GameID,
			"PublicID":        uuid.NewV4().String(),
			"OwnerID":         owner.ID,
			"Metadata":        map[string]interface{}{"x": "a"},
			"MembershipCount": approvedMemberships + 1,
		}).(*Clan)
		err = db.Insert(clan)
		if err != nil {
			return nil, err
		}

		return clan, nil
	}

	createMembership := func(approved, denied, banned bool) (*Membership, error) {
		clan, err := createClan()
		if err != nil {
			return nil, err
		}

		payload := map[string]interface{}{
			"GameID":      owner.GameID,
			"PlayerID":    player.ID,
			"ClanID":      clan.ID,
			"RequestorID": owner.ID,
			"Metadata":    map[string]interface{}{"x": "a"},
			"Approved":    approved,
			"Denied":      denied,
			"Banned":      banned,
		}
		if banned {
			payload["DeletedAt"] = util.NowMilli()
			payload["DeletedBy"] = owner.ID
		}
		if approved {
			payload["ApproverID"] = sql.NullInt64{Int64: int64(player.ID), Valid: true}
			payload["ApprovedAt"] = util.NowMilli()
		} else if denied {
			payload["DenierID"] = sql.NullInt64{Int64: int64(player.ID), Valid: true}
			payload["DeniedAt"] = util.NowMilli()
		}

		membership := MembershipFactory.MustCreateWithOption(payload).(*Membership)

		err = db.Insert(membership)
		if err != nil {
			return nil, err
		}

		return membership, nil
	}

	for i := 0; i < approvedMemberships; i++ {
		_, err := createMembership(true, false, false)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < rejectedMemberships; i++ {
		_, err := createMembership(false, true, false)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < bannedMemberships; i++ {
		_, err := createMembership(false, false, true)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < pendingMemberships; i++ {
		_, err := createMembership(false, false, false)
		if err != nil {
			return nil, err
		}
	}

	return player, nil
}
