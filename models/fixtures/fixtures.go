// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package fixtures

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/jrallison/go-workers"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	"github.com/topfreegames/khan/models"
	"github.com/topfreegames/khan/mongo"
	"github.com/topfreegames/khan/queues"
	kt "github.com/topfreegames/khan/testing"
	"github.com/topfreegames/khan/util"
)

// GameFactory is responsible for constructing test game instances
var GameFactory = factory.NewFactory(
	&models.Game{
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
		CooldownBeforeApply:           3600,
		CooldownBeforeInvite:          0,
		MaxPendingInvites:             20,
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
	&models.Hook{EventType: models.GameUpdatedHook, URL: "http://test/game-created"},
)

// CreateHookFactory is responsible for creating a test hook instance with the associated game
func CreateHookFactory(db models.DB, gameID string, eventType int, url string) (*models.Hook, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*models.Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	hook := HookFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":    gameID,
		"PublicID":  uuid.NewV4().String(),
		"EventType": eventType,
		"URL":       url,
	}).(*models.Hook)
	err = db.Insert(hook)
	if err != nil {
		return nil, err
	}
	return hook, nil
}

// GetHooksForRoutes gets hooks for all the specified routes
func GetHooksForRoutes(db models.DB, routes []string, eventType int) ([]*models.Hook, error) {
	var hooks []*models.Hook

	gameID := uuid.NewV4().String()

	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*models.Game)
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
		}).(*models.Hook)
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

var testKey []byte = []byte("00000000000000000000000000000000")

//GetEncryptionKey returns the models test encryptionKey
func GetEncryptionKey() []byte {
	return testKey
}

// PlayerFactory is responsible for constructing test player instances
var PlayerFactory = configureFactory(factory.NewFactory(
	&models.Player{},
))

// ClanFactory is responsible for constructing test clan instances
var ClanFactory = configureFactory(factory.NewFactory(
	&models.Clan{},
))

// CreatePlayerFactory is responsible for creating a test player instance with the associated game
func CreatePlayerFactory(db models.DB, gameID string, skipCreateGame ...bool) (*models.Game, *models.Player, error) {
	var game *models.Game

	if skipCreateGame == nil || len(skipCreateGame) != 1 || !skipCreateGame[0] {
		if gameID == "" {
			gameID = uuid.NewV4().String()
		}
		game = GameFactory.MustCreateWithOption(map[string]interface{}{
			"PublicID": gameID,
		}).(*models.Game)
		err := db.Insert(game)
		if err != nil {
			return nil, nil, err
		}
	} else {
		var err error
		game, err = models.GetGameByPublicID(db, gameID)
		if err != nil {
			return nil, nil, err
		}
	}

	player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID": gameID,
	}).(*models.Player)
	err := db.Insert(player)
	if err != nil {
		return nil, nil, err
	}
	return game, player, nil
}

// MembershipFactory is responsible for constructing test membership instances
var MembershipFactory = factory.NewFactory(
	&models.Membership{},
).SeqInt("GameID", func(n int) (interface{}, error) {
	return fmt.Sprintf("game-%d", n), nil
}).Attr("ApproverID", func(args factory.Args) (interface{}, error) {
	membership := args.Instance().(*models.Membership)
	approverID := int64(0)
	valid := false
	if membership.Approved {
		approverID = membership.RequestorID
		valid = true
	}
	return sql.NullInt64{Int64: int64(approverID), Valid: valid}, nil
}).Attr("DenierID", func(args factory.Args) (interface{}, error) {
	membership := args.Instance().(*models.Membership)
	denierID := int64(0)
	valid := false
	if membership.Denied {
		denierID = membership.RequestorID
		valid = true
	}
	return sql.NullInt64{Int64: denierID, Valid: valid}, nil
})

// GetClanWithMemberships returns a clan filled with the number of memberships specified
func GetClanWithMemberships(
	db models.DB, approvedMemberships, deniedMemberships, bannedMemberships, pendingMemberships int, gameID string, clanPublicID string, options ...bool) (*models.Game, *models.Clan, *models.Player, []*models.Player, []*models.Membership, error) {
	var game *models.Game

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
		}).(*models.Game)
		err := db.Insert(game)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	} else {
		var err error
		game, err = models.GetGameByPublicID(db, gameID)
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
	}).(*models.Player)
	err := db.Insert(owner)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":           owner.GameID,
		"PublicID":         clanPublicID,
		"OwnerID":          owner.ID,
		"Metadata":         map[string]interface{}{"x": "a"},
		"MembershipCount":  approvedMemberships + 1,
		"AllowApplication": true,
	}).(*models.Clan)
	err = db.Insert(clan)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	testData := []map[string]interface{}{
		{
			"approved": true,
			"denied":   false,
			"banned":   false,
			"count":    approvedMemberships,
		},
		{
			"approved": false,
			"denied":   true,
			"banned":   false,
			"count":    deniedMemberships,
		},
		{
			"approved": false,
			"denied":   false,
			"banned":   true,
			"count":    bannedMemberships,
		},
		{
			"approved": false,
			"denied":   false,
			"banned":   false,
			"count":    pendingMemberships,
		},
	}

	var players []*models.Player
	var memberships []*models.Membership

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
			}).(*models.Player)
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
			}).(*models.Membership)

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

			if options != nil && len(options) > 2 && options[2] {
				player.Metadata = map[string]interface{}{
					"id": membership.ID,
				}
				_, err = db.Update(player)
				if err != nil {
					return nil, nil, nil, nil, nil, err
				}
			}

			memberships = append(memberships, membership)
		}
	}

	return game, clan, owner, players, memberships, nil
}

// GetClanReachedMaxMemberships returns a clan with one approved membership, one unapproved membership and game MaxMembers=1
func GetClanReachedMaxMemberships(db models.DB) (*models.Game, *models.Clan, *models.Player, []*models.Player, []*models.Membership, error) {
	gameID := uuid.NewV4().String()
	clanPublicID := uuid.NewV4().String()

	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID":   gameID,
		"MaxMembers": 1,
	}).(*models.Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":         gameID,
		"OwnershipCount": 1,
	}).(*models.Player)
	err = db.Insert(owner)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var players []*models.Player

	for i := 0; i < 2; i++ {
		membershipCount := 0
		if i == 0 {
			membershipCount = 1
		}
		player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
			"GameID":          owner.GameID,
			"MembershipCount": membershipCount,
		}).(*models.Player)
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
	}).(*models.Clan)
	err = db.Insert(clan)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var memberships []*models.Membership

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
	}).(*models.Membership)
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
	}).(*models.Membership)
	err = db.Insert(membership)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	memberships = append(memberships, membership)

	return game, clan, owner, players, memberships, nil
}

// GetTestClanWithRandomPublicIDAndName returns a clan with random UUID v4 publicID and name for tests
func GetTestClanWithRandomPublicIDAndName(db models.DB, gameID string, ownerID int64) (*models.Clan, error) {
	clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":   gameID,
		"PublicID": uuid.NewV4().String(),
		"Name":     uuid.NewV4().String(),
		"OwnerID":  ownerID,
	}).(*models.Clan)
	err := db.Insert(clan)
	if err != nil {
		return nil, err
	}
	err = clan.UpdateClanIntoMongoDB()
	if err != nil {
		return nil, err
	}
	return clan, nil
}

// GetTestClanWithName returns a clan with name for tests
func GetTestClanWithName(db models.DB, gameID, name string, ownerID int64) (*models.Clan, error) {
	clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":   gameID,
		"PublicID": uuid.NewV4().String(),
		"Name":     name,
		"OwnerID":  ownerID,
	}).(*models.Clan)
	err := db.Insert(clan)
	if err != nil {
		return nil, err
	}
	err = clan.UpdateClanIntoMongoDB()
	if err != nil {
		return nil, err
	}
	return clan, nil
}

// AfterClanCreationHook permits a caller to execute code after a test clan is created.
type AfterClanCreationHook func(player *models.Player, clan *models.Clan) error

// EnqueueClanForMongoUpdate is a possible value for the type AfterClanCreationHook. This is will enqueue
// the clan into Redis for the mongo worker to insert/update it on MongoDB
func EnqueueClanForMongoUpdate(player *models.Player, clan *models.Clan) error {
	return clan.UpdateClanIntoMongoDB()
}

// CreateTestClans returns a list of clans for tests
func CreateTestClans(
	db models.DB, mongoDB interfaces.MongoDB, gameID string, publicIDTemplate string, numberOfClans int, afterCreationHook AfterClanCreationHook,
) (*models.Player, []*models.Clan, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*models.Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, err
	}

	player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":         gameID,
		"OwnershipCount": numberOfClans,
	}).(*models.Player)
	err = db.Insert(player)
	if err != nil {
		return nil, nil, err
	}

	if publicIDTemplate == "" {
		publicIDTemplate = uuid.NewV4().String()
	}

	var clans []*models.Clan
	for i := 0; i < numberOfClans; i++ {
		clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
			"GameID":   player.GameID,
			"PublicID": fmt.Sprintf("%s-%d", publicIDTemplate, i),
			"Name":     fmt.Sprintf("💩clán-%s-%d", publicIDTemplate, i),
			"OwnerID":  player.ID,
		}).(*models.Clan)
		err = db.Insert(clan)
		if err != nil {
			return nil, nil, err
		}
		err = afterCreationHook(player, clan)
		if err != nil {
			return nil, nil, err
		}

		clans = append(clans, clan)
	}

	err = mongoDB.Run(mongo.GetClanNameTextIndexCommand(gameID, false), nil)
	if err != nil {
		return nil, nil, err
	}

	return player, clans, nil
}

// GetTestHooks return a fixed number of hooks for each event available
func GetTestHooks(db models.DB, gameID string, numberOfHooks int) ([]*models.Hook, error) {
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*models.Game)
	err := db.Insert(game)
	if err != nil {
		return nil, err
	}

	var hooks []*models.Hook
	for i := 0; i < 2; i++ {
		for j := 0; j < numberOfHooks; j++ {
			hook := HookFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":    gameID,
				"PublicID":  uuid.NewV4().String(),
				"EventType": i,
				"URL":       fmt.Sprintf("http://test/event-%d-%d", i, j),
			}).(*models.Hook)
			err = db.Insert(hook)
			if err != nil {
				return nil, err
			}
			hooks = append(hooks, hook)
		}
	}

	return hooks, nil
}

//GetTestPlayerWithMemberships returns the clan owner, a player with approved, rejected, and banned memberships
func GetTestPlayerWithMemberships(db models.DB, gameID string, approvedMemberships, rejectedMemberships, bannedMemberships, pendingMemberships int) (*models.Player, *models.Player, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
	}).(*models.Game)
	err := db.Insert(game)
	if err != nil {
		return nil, nil, err
	}

	owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":         gameID,
		"OwnershipCount": 1,
	}).(*models.Player)
	err = db.Insert(owner)
	if err != nil {
		return nil, nil, err
	}

	player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":          owner.GameID,
		"MembershipCount": approvedMemberships,
	}).(*models.Player)
	err = db.Insert(player)
	if err != nil {
		return nil, nil, err
	}

	createClan := func() (*models.Clan, error) {
		clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
			"GameID":          owner.GameID,
			"PublicID":        uuid.NewV4().String(),
			"OwnerID":         owner.ID,
			"Metadata":        map[string]interface{}{"x": "a"},
			"MembershipCount": approvedMemberships + 1,
		}).(*models.Clan)
		err = db.Insert(clan)
		if err != nil {
			return nil, err
		}

		return clan, nil
	}

	createMembership := func(approved, denied, banned bool) (*models.Membership, error) {
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

		membership := MembershipFactory.MustCreateWithOption(payload).(*models.Membership)

		err = db.Insert(membership)
		if err != nil {
			return nil, err
		}

		return membership, nil
	}

	for i := 0; i < approvedMemberships; i++ {
		_, err := createMembership(true, false, false)
		if err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < rejectedMemberships; i++ {
		_, err := createMembership(false, true, false)
		if err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < bannedMemberships; i++ {
		_, err := createMembership(false, false, true)
		if err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < pendingMemberships; i++ {
		_, err := createMembership(false, false, false)
		if err != nil {
			return nil, nil, err
		}
	}

	return owner, player, nil
}

//GetTestClanWithStaleData returns a player with approved, rejected and banned memberships
func GetTestClanWithStaleData(db models.DB, staleApplications, staleInvites, staleDenies, staleDeletes int) (string, error) {
	gameID := uuid.NewV4().String()
	game := GameFactory.MustCreateWithOption(map[string]interface{}{
		"PublicID": gameID,
		"Metadata": map[string]interface{}{
			"pendingApplicationsExpiration": 3600,
			"pendingInvitesExpiration":      3600,
			"deniedMembershipsExpiration":   3600,
			"deletedMembershipsExpiration":  3600,
		},
	}).(*models.Game)
	err := db.Insert(game)
	if err != nil {
		return "", err
	}

	owner := PlayerFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":         gameID,
		"OwnershipCount": 1,
	}).(*models.Player)
	err = db.Insert(owner)
	if err != nil {
		return "", err
	}

	clan := ClanFactory.MustCreateWithOption(map[string]interface{}{
		"GameID":   owner.GameID,
		"PublicID": uuid.NewV4().String(),
		"OwnerID":  owner.ID,
		"Metadata": map[string]interface{}{"x": "a"},
	}).(*models.Clan)
	err = db.Insert(clan)
	if err != nil {
		return "", err
	}

	createMembership := func(createdAt int64, application bool, denied bool, deleted bool) (*models.Player, *models.Membership, error) {
		player := PlayerFactory.MustCreateWithOption(map[string]interface{}{
			"GameID": gameID,
		}).(*models.Player)
		err = db.Insert(player)
		if err != nil {
			return nil, nil, err
		}

		payload := map[string]interface{}{
			"GameID":      owner.GameID,
			"PlayerID":    player.ID,
			"ClanID":      clan.ID,
			"RequestorID": player.ID,
			"Metadata":    map[string]interface{}{"x": "a"},
			"Approved":    false,
			"Denied":      false,
			"Banned":      false,
			"CreatedAt":   createdAt,
			"UpdatedAt":   createdAt,
		}

		if !application {
			payload["RequestorID"] = owner.ID
		}

		if denied {
			payload["Denied"] = true
		}

		if deleted {
			payload["DeletedAt"] = util.NowMilli()
		}

		membership := MembershipFactory.MustCreateWithOption(payload).(*models.Membership)
		err = db.Insert(membership)
		if err != nil {
			return nil, nil, err
		}
		_, err := db.Exec(`UPDATE memberships SET updated_at=$1, created_at=$1 WHERE id=$2`, createdAt, membership.ID)
		if err != nil {
			return nil, nil, err
		}
		return player, membership, nil
	}

	for i := 0; i < staleApplications*2; i++ {
		createMembership(util.NowMilli(), true, false, false)
	}

	for i := 0; i < staleInvites*2; i++ {
		createMembership(util.NowMilli(), false, false, false)
	}

	for i := 0; i < staleDenies*2; i++ {
		createMembership(util.NowMilli(), false, true, false)
	}

	for i := 0; i < staleDeletes*2; i++ {
		createMembership(util.NowMilli(), false, false, true)
	}

	expiration := int64((time.Duration(600) * time.Hour).Seconds() * 1000)
	for i := 0; i < staleApplications; i++ {
		createMembership(util.NowMilli()-expiration, true, false, false)
	}

	for i := 0; i < staleInvites; i++ {
		createMembership(util.NowMilli()-expiration, false, false, false)
	}

	for i := 0; i < staleDenies; i++ {
		createMembership(util.NowMilli()-expiration, false, true, false)
	}

	for i := 0; i < staleDeletes; i++ {
		createMembership(util.NowMilli()-expiration, false, false, true)
	}

	return gameID, nil
}

// ConfigureAndStartGoWorkers starts the mongo workers
func ConfigureAndStartGoWorkers() (*models.MongoWorker, error) {
	config := viper.New()
	config.SetConfigType("yaml")
	config.SetConfigFile("../config/test.yaml")
	err := config.ReadInConfig()
	if err != nil {
		return nil, err
	}

	redisHost := config.GetString("redis.host")
	redisPort := config.GetInt("redis.port")
	redisDatabase := config.GetInt("redis.database")
	redisPool := config.GetInt("redis.pool")
	workerCount := config.GetInt("webhooks.workers")
	if redisPool == 0 {
		redisPool = 30
	}

	if workerCount == 0 {
		workerCount = 5
	}

	opts := map[string]string{
		// location of redis instance
		"server": fmt.Sprintf("%s:%d", redisHost, redisPort),
		// instance of the database
		"database": strconv.Itoa(redisDatabase),
		// number of connections to keep open with redis
		"pool": strconv.Itoa(redisPool),
		// unique process id
		"process": uuid.NewV4().String(),
	}
	redisPass := config.GetString("redis.password")
	if redisPass != "" {
		opts["password"] = redisPass
	}
	workers.Configure(opts)

	logger := kt.NewMockLogger()
	mongoWorker := models.NewMongoWorker(logger, config)
	workers.Process(queues.KhanMongoQueue, mongoWorker.PerformUpdateMongo, workerCount)
	workers.Start()
	return mongoWorker, nil
}
