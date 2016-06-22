// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import "database/sql"

type clanDetailsDAO struct {
	// Clan general information
	GameID               string
	ClanPublicID         string
	ClanName             string
	ClanMetadata         string
	ClanAllowApplication bool
	ClanAutoJoin         bool

	//Membership Information
	MembershipLevel     sql.NullInt64
	MembershipApproved  sql.NullBool
	MembershipDenied    sql.NullBool
	MembershipCreatedAt sql.NullInt64
	MembershipUpdatedAt sql.NullInt64

	// Clan Owner Information
	OwnerPublicID string
	OwnerName     string
	OwnerMetadata string

	// Member Information
	PlayerPublicID sql.NullString
	PlayerName     sql.NullString
	PlayerMetadata sql.NullString

	// Requestor Information
	RequestorPublicID sql.NullString
	RequestorName     sql.NullString
}

func (member *clanDetailsDAO) Serialize() map[string]interface{} {
	return map[string]interface{}{
		// No need to include clan information as that will be available in the payload already
		"membershipLevel":     nullOrInt(member.MembershipLevel),
		"membershipApproved":  nullOrBool(member.MembershipApproved),
		"membershipDenied":    nullOrBool(member.MembershipDenied),
		"membershipCreatedAt": nullOrInt(member.MembershipCreatedAt),
		"membershipUpdatedAt": nullOrInt(member.MembershipUpdatedAt),
		"playerPublicID":      nullOrString(member.PlayerPublicID),
		"playerName":          nullOrString(member.PlayerName),
		"playerMetadata":      nullOrString(member.PlayerMetadata),
		"requestorPublicID":   nullOrString(member.RequestorPublicID),
		"requestorName":       nullOrString(member.RequestorName),
	}
}
