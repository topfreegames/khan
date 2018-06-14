package lib

import "fmt"

type requestError struct {
	statusCode int
	body       string
}

func newRequestError(statusCode int, body string) *requestError {
	return &requestError{
		statusCode: statusCode,
		body:       body,
	}
}

func (r *requestError) Error() string {
	return fmt.Sprintf("Request error. Status code: %d. Body: %s", r.statusCode, r.body)
}

//ClanPayload maps the payload for the Create Clan route and Update Clan route
type ClanPayload struct {
	PublicID         string                 `json:"publicID,omitempty"`
	Name             string                 `json:"name"`
	OwnerPublicID    string                 `json:"ownerPublicID"`
	Metadata         map[string]interface{} `json:"metadata"`
	AllowApplication bool                   `json:"allowApplication"`
	AutoJoin         bool                   `json:"autoJoin"`
}

// Player defines the struct returned by the khan API for retrieve player
type Player struct {
	PublicID    string              `json:"publicID"`
	Name        string              `json:"name"`
	Metadata    interface{}         `json:"metadata"`
	Clans       *ClansRelationships `json:"clans,omitempty"`
	Memberships []*Membership       `json:"memberships,omitempty"`
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

// Membership defines the membership returned by retrieve player
type Membership struct {
	Approved   bool                     `json:"approved"`
	Banned     bool                     `json:"banned"`
	Denied     bool                     `json:"denied"`
	Clan       *ClanPlayerInfo          `json:"clan"`
	CreatedAt  int64                    `json:"createdAt"`
	UpdatedAt  int64                    `json:"updatedAt"`
	DeletedAt  int64                    `json:"deletedAt"`
	ApprovedAt int64                    `json:"approvedAt"`
	DeniedAt   int64                    `json:"deniedAt"`
	Level      string                   `json:"level"`
	Message    string                   `json:"message"`
	Requestor  *ApproverRequestorDenier `json:"requestor"`
	Approver   *ApproverRequestorDenier `json:"approver"`
	Denier     *ApproverRequestorDenier `json:"denier"`
}

// ClanPlayerInfo defines the clan info returned on the membership
type ClanPlayerInfo struct {
	Metadata        interface{} `json:"metadata"`
	Name            string      `json:"name"`
	PublicID        string      `json:"publicID"`
	MembershipCount int         `json:"membershipCount"`
}

// ApproverRequestorDenier defines the data returned for these elements on each membership
type ApproverRequestorDenier struct {
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
