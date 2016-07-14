// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/topfreegames/khan/models"
)

var currentStage int
var stages map[int]string

func createGames(db models.DB, gameCount int) []string {
	sql := `INSERT INTO games(
		public_id,
		name,
		min_membership_level,
		max_membership_level,
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
		cooldown_after_deny,
		cooldown_after_delete,
		created_at,
		updated_at) SELECT
			uuid_generate_v4(),
			uuid_generate_v4(),
			1,
			3,
			1,
			1,
			1,
			1,
			1,
			1,
			50,
			1, 
			CAST(to_jsonb($1::text) as jsonb),
			CAST(to_jsonb($2::text) as jsonb),
			0,
			0,
			0,
			0
	FROM generate_series(1, $3)
	`
	_, err := db.Exec(sql, "{1:\"member\", 2:\"elder\", 3:\"coleader\"}", "{}", gameCount)

	if err != nil {
		panic(err.Error())
	}

	var gameIDs []string
	_, err = db.Select(&gameIDs, "select public_id from games")
	if err != nil {
		panic(err.Error())
	}

	return gameIDs
}

func createPlayersWithoutClan(db models.DB, games []string, playersWithoutClan int, progress func() bool) {
	for _, game := range games {
		sql := `
		INSERT INTO players(
			public_id,
			game_id,
			name,
			metadata,
			created_at,
			updated_at
		) SELECT
				uuid_generate_v4(),
				$1,
				uuid_generate_v4(),
				$2,
				0,
				0
		FROM generate_series(1, $3)
		`
		_, err := db.Exec(sql, game, "{}", playersWithoutClan)
		if err != nil {
			panic(err.Error())
		}
		progress()
	}
}

type clanData struct {
	ID       int
	GameID   string
	PublicID string
	OwnerID  int
}

func createClans(db models.DB, games []string, clans int, progress func() bool) map[string][]clanData {
	for _, game := range games {
		sql := `
		WITH owner AS (
			INSERT INTO players(
				public_id,
				game_id,
				name,
				metadata,
				created_at,
				updated_at
			) SELECT
					uuid_generate_v4(),
					$1,
					uuid_generate_v4(),
					$2,
					0,
					0
			FROM generate_series(1, 1)
			RETURNING *
		)

		INSERT INTO clans(
			public_id,
			game_id,
			name,
			metadata,
			allow_application,
			auto_join,
			created_at,
			updated_at,
			deleted_at,
			owner_id
		) SELECT
			uuid_generate_v4(),
			$1,
			uuid_generate_v4(),
			$2,
			true,
			true,
			0,
			0,
			0,
			owner.id
		FROM generate_series(1, $3), owner
		`
		_, err := db.Exec(sql, game, "{}", clans)
		if err != nil {
			panic(err.Error())
		}

		progress()
	}

	var allClans []clanData
	_, err := db.Select(&allClans, "select id ID, game_id GameID, public_id PublicID, owner_id OwnerID from clans")
	if err != nil {
		panic(err.Error())
	}

	clanMap := map[string][]clanData{}

	for _, clan := range allClans {
		clanMap[clan.GameID] = append(clanMap[clan.GameID], clan)
	}

	return clanMap
}

func createClanPlayers(db models.DB, games []string, clans map[string][]clanData, playersPerClan int, approved, denied, banned bool, progress func() bool) {
	for _, game := range games {
		for _, clan := range clans[game] {
			sql := `
			WITH addedPlayers AS (
				INSERT INTO players(
					public_id,
					game_id,
					name,
					metadata,
					created_at,
					updated_at
				) SELECT
						uuid_generate_v4(),
						$1,
						uuid_generate_v4(),
						$8,
						0,
						0
				FROM generate_series(1, $2)
				RETURNING *
			)

			INSERT INTO memberships(
				game_id,
				clan_id,
				player_id,
				membership_level,
				approved,
				denied,
				banned,
				requestor_id,
				created_at,
				updated_at,
				deleted_by,
				deleted_at
			) SELECT
				ap.game_id,
				$3,
				ap.id,
				'member',
				$4,
				$5,
				$6,
				$7,
				0,
				0,
				null,
				0
			FROM addedPlayers ap
			`
			_, err := db.Exec(sql, game, playersPerClan, clan.ID, approved, denied, banned, clan.OwnerID, "{}")
			if err != nil {
				panic(err.Error())
			}

			progress()
		}
	}
}

func createTestData(db models.DB, games, clansPerGame, playersPerClan, playersWithoutClan, pendingMembershipsPerClan, deniedMembershipsPerClan, bannedMembershipsPerClan int, progress func() bool) error {
	gameIDs := createGames(db, games)
	progress()
	currentStage++

	//fmt.Println("Creating players without clan...")
	createPlayersWithoutClan(db, gameIDs, playersWithoutClan, progress)
	currentStage++

	//fmt.Println("Creating clans...")
	clans := createClans(db, gameIDs, clansPerGame, progress)
	currentStage++

	//fmt.Println("Creating players with approved membership...")
	createClanPlayers(db, gameIDs, clans, playersPerClan, true, false, false, progress)
	currentStage++

	//Pending memberships
	//fmt.Println("Creating players with pending membership...")
	createClanPlayers(db, gameIDs, clans, pendingMembershipsPerClan, false, false, false, progress)
	currentStage++

	//Denied memberships
	//fmt.Println("Creating players with denied membership...")
	createClanPlayers(db, gameIDs, clans, deniedMembershipsPerClan, false, true, false, progress)
	currentStage++

	//Banned memberships
	//fmt.Println("Creating players with banned membership...")
	createClanPlayers(db, gameIDs, clans, bannedMembershipsPerClan, false, false, true, progress)

	return nil
}

var games = flag.Int("games", 20, "number of games to create")
var playersWithoutClan = flag.Int("pwc", 50000, "number of players without clan")
var clansPerGame = flag.Int("cpg", 1000, "clans per game")
var playersPerClan = flag.Int("ppc", 50, "number of players in each clan")
var pendingMembershipsPerClan = flag.Int("pmpc", 250, "number of players with pending memberships in each clan")
var bannedMembershipsPerClan = flag.Int("bmpc", 250, "number of players with pending memberships in each clan")
var deniedMembershipsPerClan = flag.Int("dmpc", 250, "number of players with pending memberships in each clan")
var useMainDB = flag.Bool("use-main", false, "use main database for khan")

func main() {
	flag.Parse()
	stages = map[int]string{
		0:  "******Games*****",
		1:  "Clanless Players",
		2:  "******Clans*****",
		3:  "Approved Members",
		4:  "*Pending Members",
		5:  "*Denied Members*",
		6:  "*Banned Members*",
		7:  "*Banned Members*",
		8:  "*Banned Members*",
		9:  "*Banned Members*",
		10: "*Banned Members*",
	}

	start := time.Now()

	totalOps := 1 + *games + *games + (*games * *clansPerGame * 4)

	uiprogress.Start()                     // start rendering
	bar := uiprogress.AddBar(totalOps - 1) // Add a new bar
	bar.AppendCompleted()
	bar.PrependElapsed()

	// prepend the deploy step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		ellapsed := time.Now().Sub(start)
		itemsPerSec := float64(b.Current()+1) / ellapsed.Seconds()
		timeToComplete := float64(totalOps) / itemsPerSec / 60.0 / 60.0
		text := fmt.Sprintf("[%s] %d/%d (%.2fhs to complete)", stages[currentStage], b.Current()+1, totalOps, timeToComplete)
		return strutil.Resize(text, uint(len(text)))
	})

	var testDb models.DB
	var err error

	if *useMainDB {
		testDb, err = models.GetDefaultDB()
	} else {
		testDb, err = models.GetPerfDB()
	}
	if err != nil {
		panic(err.Error())
	}

	createTestData(testDb, *games, *clansPerGame, *playersPerClan, *playersWithoutClan, *pendingMembershipsPerClan, *deniedMembershipsPerClan, *bannedMembershipsPerClan, bar.Incr)
}
