// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package models

import "database/sql"

type clanDetailsDAO struct {
	GameID              string
	ClanPublicID        string
	ClanName            string
	ClanMetadata        string
	MembershipLevel     sql.NullInt64
	MembershipApproved  sql.NullBool
	MembershipDenied    sql.NullBool
	MembershipCreatedAt sql.NullInt64
	MembershipUpdatedAt sql.NullInt64
	PlayerPublicID      sql.NullString
	PlayerName          sql.NullString
	PlayerMetadata      sql.NullString
	RequestorPublicID   sql.NullString
	RequestorName       sql.NullString
}

func (member *clanDetailsDAO) Serialize() map[string]interface{} {
	return map[string]interface{}{
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
