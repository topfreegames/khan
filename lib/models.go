package lib

import "fmt"

// RequestError contains code and body of a request that failed
type RequestError struct {
	statusCode int
	body       string
}

func newRequestError(statusCode int, body string) *RequestError {
	return &RequestError{
		statusCode: statusCode,
		body:       body,
	}
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("Request error. Status code: %d. Body: %s", r.statusCode, r.body)
}

//Status returns the status code of the error
func (r *RequestError) Status() int {
	return r.statusCode
}

//ClanPayload maps the payload for the Create Clan route and Update Clan route
type ClanPayload struct {
	PublicID         string      `json:"publicID,omitempty"`
	Name             string      `json:"name"`
	OwnerPublicID    string      `json:"ownerPublicID"`
	Metadata         interface{} `json:"metadata"`
	AllowApplication bool        `json:"allowApplication"`
	AutoJoin         bool        `json:"autoJoin"`
}

// Player defines the struct returned by the khan API for retrieve player
type Player struct {
	PublicID    string              `json:"publicID"`
	Name        string              `json:"name"`
	Metadata    interface{}         `json:"metadata"`
	Clans       *ClansRelationships `json:"clans,omitempty"`
	Memberships []*PlayerMembership `json:"memberships,omitempty"`
}

// ClansRelationships defines the struct returned inside player
type ClansRelationships struct {
	Owned               []*ClanNameAndPublicID `json:"owned"`
	Approved            []*ClanNameAndPublicID `json:"approved"`
	Banned              []*ClanNameAndPublicID `json:"banned"`
	Denied              []*ClanNameAndPublicID `json:"denied"`
	PendingApplications []*ClanNameAndPublicID `json:"pendingApplications"`
	PendingInvites      []*ClanNameAndPublicID `json:"pendingInvites"`
}

// ClanNameAndPublicID has name and publicID
type ClanNameAndPublicID struct {
	Name     string `json:"name"`
	PublicID string `json:"publicID"`
}

// PlayerMembership defines the membership returned by retrieve player
type PlayerMembership struct {
	Approved   bool             `json:"approved"`
	Banned     bool             `json:"banned"`
	Denied     bool             `json:"denied"`
	Clan       *ClanPlayerInfo  `json:"clan"`
	CreatedAt  int64            `json:"createdAt"`
	UpdatedAt  int64            `json:"updatedAt"`
	DeletedAt  int64            `json:"deletedAt"`
	ApprovedAt int64            `json:"approvedAt"`
	DeniedAt   int64            `json:"deniedAt"`
	Level      string           `json:"level"`
	Message    string           `json:"message"`
	Requestor  *ShortPlayerInfo `json:"requestor"`
	Approver   *ShortPlayerInfo `json:"approver"`
	Denier     *ShortPlayerInfo `json:"denier"`
}

// ClanPlayerInfo defines the clan info returned on the membership
type ClanPlayerInfo struct {
	Metadata        interface{} `json:"metadata"`
	Name            string      `json:"name"`
	PublicID        string      `json:"publicID"`
	MembershipCount int         `json:"membershipCount"`
}

// ShortPlayerInfo defines the data returned for these elements on each membership
type ShortPlayerInfo struct {
	PublicID string      `json:"publicID"`
	Name     string      `json:"name"`
	Metadata interface{} `json:"metadata"`
}

// ClanSummary defines the clan summary
type ClanSummary struct {
	PublicID         string      `json:"publicID"`
	Name             string      `json:"name"`
	Metadata         interface{} `json:"metadata"`
	AllowApplication bool        `json:"allowApplication"`
	AutoJoin         bool        `json:"autoJoin"`
	MembershipCount  int         `json:"membershipCount"`
}

// ClansSummary defines the clans summary
type ClansSummary struct {
	Clans []*ClanSummary `json:"clans"`
}

// ClanMembershipPlayer represents the player structure inside the clan membership
type ClanMembershipPlayer struct {
	Approver *ShortPlayerInfo `json:"approver"`
	Metadata interface{}      `json:"metadata"`
	Name     string           `json:"name"`
	PublicID string           `json:"publicID"`
}

// ClanMembership represents the membership structure inside a clan response
type ClanMembership struct {
	Level   int                   `json:"level"`
	Message string                `json:"message"`
	Player  *ClanMembershipPlayer `json:"player"`
}

// ClanMemberships is the memberships structure inside a clan response
type ClanMemberships struct {
	PendingApplications []*ClanMembership `json:"pendingApplications"`
	PendingInvites      []*ClanMembership `json:"pendingInvites"`
	Denied              []*ClanMembership `json:"denied"`
	Banned              []*ClanMembership `json:"banned"`
}

// Clan is the structure returned by the retrieve clan route
type Clan struct {
	PublicID         string            `json:"publicID"`
	Name             string            `json:"name"`
	Metadata         interface{}       `json:"metadata"`
	AllowApplication bool              `json:"allowApplication"`
	AutoJoin         bool              `json:"autoJoin"`
	MembershipCount  int               `json:"membershipCount"`
	Owner            *ShortPlayerInfo  `json:"owner"`
	Roster           []*ClanMembership `json:"roster"`
	Memberships      *ClanMemberships  `json:"memberships"`
}
