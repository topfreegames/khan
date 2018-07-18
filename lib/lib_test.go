package lib_test

import (
	"github.com/spf13/viper"
	"github.com/topfreegames/khan/lib"
	httpmock "gopkg.in/jarcoal/httpmock.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lib", func() {
	var k lib.KhanInterface
	var config *viper.Viper
	var gameID string

	BeforeSuite(func() {
		config = viper.New()
		httpmock.Activate()
	})

	BeforeEach(func() {
		//default configs for each test
		config.Set("khan.url", "http://khan")
		config.Set("khan.user", "user")
		config.Set("khan.pass", "pass")
		config.Set("khan.gameid", "testgame")
		gameID = "testgame"
		k = lib.NewKhan(config)
		httpmock.Reset()
	})

	Describe("NewKhan", func() {
		It("Should start a new instance of Khan Lib", func() {
			k = lib.NewKhan(config)
			Expect(k).NotTo(BeNil())
		})
	})

	Describe("CreatePlayer", func() {
		It("Should call khan API to create player", func() {
			url := "http://khan/games/" + gameID + "/players"
			publicID := "testid"
			httpmock.RegisterResponder("POST", url,
				httpmock.NewStringResponder(200, `{ "success": true, "publicID": "testid" }`))

			playerID, err := k.CreatePlayer(nil, publicID, "testname", nil)

			Expect(err).To(BeNil())
			Expect(playerID).To(Equal(publicID))
		})
	})

	Describe("UpdatePlayer", func() {
		It("Should call khan API to update player", func() {
			publicID := "testid"
			url := "http://khan/games/" + gameID + "/players/" + publicID
			httpmock.RegisterResponder("PUT", url,
				httpmock.NewStringResponder(200, `{ "success": true, "publicID": "testid" }`))

			err := k.UpdatePlayer(nil, publicID, "testname", nil)

			Expect(err).To(BeNil())
		})
	})

	Describe("RetrievePlayer", func() {
		It("Should call khan API to retrieve player", func() {
			publicID := "testid"
			url := "http://khan/games/" + gameID + "/players/" + publicID
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{
					"success": true,
					"publicID": "testid",
					"name": "testname",
					"metadata": {},
					"clans": {
						"owned": [],
						"approved": [
						  { "name": "approvedClan", "publicID": "approvedClanID" }
						],
						"banned": [],
						"denied": [],
						"pendingApplications": [],
						"pendingInvites": []
					},
					"memberships": [{
						"approved": true,
						"denied": false,
						"banned": false,
						"clan": {
							"metadata": {},
							"name": "approvedClan",
							"publicID": "approvedClanID",
							"membershipCount": 2
						},
						"createdAt": 123456789,
						"updatedAt": 123456789,
						"level": "member",
						"message": "",
						"requestor": { "publicID": "testid", "name": "testname", "metadata": {} },
						"approver": {},
						"denier": {}
					}]
				}`))

			player, err := k.RetrievePlayer(nil, publicID)

			Expect(err).To(BeNil())
			Expect(player.PublicID).To(Equal(publicID))
			Expect(player.Name).To(Equal("testname"))
			Expect(player.Clans).NotTo(BeNil())
			Expect(player.Clans.Owned).To(HaveLen(0))
			Expect(player.Clans.Approved).To(HaveLen(1))
			Expect(player.Clans.Approved[0].Name).To(Equal("approvedClan"))
			Expect(player.Clans.Approved[0].PublicID).To(Equal("approvedClanID"))
			Expect(player.Memberships).To(HaveLen(1))
			Expect(player.Memberships[0].Approved).To(Equal(true))
			Expect(player.Memberships[0].Denied).To(Equal(false))
			Expect(player.Memberships[0].Banned).To(Equal(false))
			Expect(player.Memberships[0].Clan.Name).To(Equal("approvedClan"))
			Expect(player.Memberships[0].Clan.PublicID).To(Equal("approvedClanID"))
			Expect(player.Memberships[0].Clan.MembershipCount).To(Equal(2))
			Expect(player.Memberships[0].CreatedAt).To(Equal(int64(123456789)))
			Expect(player.Memberships[0].UpdatedAt).To(Equal(int64(123456789)))
			Expect(player.Memberships[0].Level).To(Equal("member"))
			Expect(player.Memberships[0].Requestor.PublicID).To(Equal(publicID))
			Expect(player.Memberships[0].Requestor.Name).To(Equal("testname"))
		})
	})

	Describe("CreateClan", func() {
		It("Should call khan API to create clan", func() {
			url := "http://khan/games/" + gameID + "/clans"
			publicID := "testid"
			httpmock.RegisterResponder("POST", url,
				httpmock.NewStringResponder(200, `{ "success": true, "publicID": "testid" }`))

			clanPayload := &lib.ClanPayload{
				PublicID:         publicID,
				Name:             "testname",
				OwnerPublicID:    "ownerID",
				AllowApplication: false,
				AutoJoin:         false,
			}
			clanID, err := k.CreateClan(nil, clanPayload)

			Expect(err).To(BeNil())
			Expect(clanID).To(Equal(publicID))
		})
	})

	Describe("UpdateClan", func() {
		It("Should call khan API to update clan", func() {
			publicID := "testid"
			url := "http://khan/games/" + gameID + "/clans/" + publicID
			httpmock.RegisterResponder("PUT", url,
				httpmock.NewStringResponder(200, `{ "success": true }`))

			clanPayload := &lib.ClanPayload{
				PublicID:         publicID,
				Name:             "testname",
				OwnerPublicID:    "ownerID",
				AllowApplication: false,
				AutoJoin:         false,
			}
			err := k.UpdateClan(nil, clanPayload)

			Expect(err).To(BeNil())
		})
	})

	Describe("RetrieveClan", func() {
		It("Should call khan API to retrieve clan", func() {
			publicID := "testid"
			url := "http://khan/games/" + gameID + "/clans/" + publicID
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{
					"success": true,
					"publicID": "testid",
					"name": "testname",
					"metadata": {},
					"allowApplication": true,
					"autoJoin": false,
					"membershipCount": 3,
					"owner": {
						"publicID": "ownerID",
						"name": "owner1"
					},
					"roster": [
					  {
						  "level": 1,
						  "message": "hey",
						  "player": {
							  "publicID": "pid1",
							  "name": "name1",
							  "approver": {
								  "publicID": "ownerID",
								  "name": "owner1"
							  }
						  }
					  },
					  {
						  "level": 1,
						  "message": "hey!",
						  "player": {
							  "publicID": "pid2",
							  "name": "name2",
							  "approver": {
								  "publicID": "ownerID",
								  "name": "owner1"
							  }
						  }
					  }
					]
				}`))

			clan, err := k.RetrieveClan(nil, publicID)

			Expect(err).To(BeNil())
			Expect(clan.PublicID).To(Equal(publicID))
			Expect(clan.Name).To(Equal("testname"))
			Expect(clan.AllowApplication).To(Equal(true))
			Expect(clan.AutoJoin).To(Equal(false))
			Expect(clan.MembershipCount).To(Equal(3))
			Expect(clan.Owner.PublicID).To(Equal("ownerID"))
			Expect(clan.Owner.Name).To(Equal("owner1"))
			Expect(clan.Roster).To(HaveLen(2))
			Expect(clan.Roster[0].Level).To(Equal(1))
			Expect(clan.Roster[0].Message).To(Equal("hey"))
			Expect(clan.Roster[0].Player.PublicID).To(Equal("pid1"))
			Expect(clan.Roster[0].Player.Name).To(Equal("name1"))
			Expect(clan.Roster[0].Player.Approver.PublicID).To(Equal("ownerID"))
			Expect(clan.Roster[0].Player.Approver.Name).To(Equal("owner1"))
			Expect(clan.Roster[1].Level).To(Equal(1))
			Expect(clan.Roster[1].Message).To(Equal("hey!"))
			Expect(clan.Roster[1].Player.PublicID).To(Equal("pid2"))
			Expect(clan.Roster[1].Player.Name).To(Equal("name2"))
			Expect(clan.Roster[1].Player.Approver.PublicID).To(Equal("ownerID"))
			Expect(clan.Roster[1].Player.Approver.Name).To(Equal("owner1"))
		})
	})

	Describe("RetrieveClanSummary", func() {
		It("Should call khan API to retrieve clan summary", func() {
			publicID := "testid"
			url := "http://khan/games/" + gameID + "/clans/" + publicID + "/summary"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{
					"success": true,
					"publicID": "testid",
					"name": "testname",
					"metadata": {},
					"allowApplication": true,
					"autoJoin": false,
					"membershipCount": 3
				}`))

			clan, err := k.RetrieveClanSummary(nil, publicID)

			Expect(err).To(BeNil())
			Expect(clan.PublicID).To(Equal(publicID))
			Expect(clan.Name).To(Equal("testname"))
			Expect(clan.AllowApplication).To(Equal(true))
			Expect(clan.AutoJoin).To(Equal(false))
			Expect(clan.MembershipCount).To(Equal(3))
		})
	})

	Describe("RetrieveClansSummary", func() {
		It("Should call khan API to retrieve clans summary", func() {
			publicIDs := []string{"testid", "testid2"}
			url := "http://khan/games/" + gameID + "/clans-summary?clanPublicIds=testid,testid2"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{
					"success": true,
					"clans": [
						{
							"publicID": "testid",
							"name": "testname",
							"metadata": {},
							"allowApplication": true,
							"autoJoin": false,
							"membershipCount": 3
						},
						{
							"publicID": "testid2",
							"name": "testname2",
							"metadata": {},
							"allowApplication": false,
							"autoJoin": false,
							"membershipCount": 1
						}
					]
				}`))

			clans, err := k.RetrieveClansSummary(nil, publicIDs)

			Expect(err).To(BeNil())
			Expect(clans).To(HaveLen(2))
			Expect(clans[0].PublicID).To(Equal(publicIDs[0]))
			Expect(clans[0].Name).To(Equal("testname"))
			Expect(clans[0].AllowApplication).To(Equal(true))
			Expect(clans[0].AutoJoin).To(Equal(false))
			Expect(clans[0].MembershipCount).To(Equal(3))
			Expect(clans[1].PublicID).To(Equal(publicIDs[1]))
			Expect(clans[1].Name).To(Equal("testname2"))
			Expect(clans[1].AllowApplication).To(Equal(false))
			Expect(clans[1].AutoJoin).To(Equal(false))
			Expect(clans[1].MembershipCount).To(Equal(1))
		})
	})

	AfterSuite(func() {
		defer httpmock.DeactivateAndReset()
	})
})
