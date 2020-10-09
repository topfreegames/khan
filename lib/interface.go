package lib

import "context"

// KhanInterface defines the interface for the khan client
type KhanInterface interface {
	ApplyForMembership(context.Context, *ApplicationPayload) (*ClanApplyResult, error)
	ApproveDenyMembershipApplication(context.Context, *ApplicationApprovalPayload) (*Result, error)
	ApproveDenyMembershipInvitation(context.Context, *InvitationApprovalPayload) (*Result, error)
	CreateClan(context.Context, *ClanPayload) (string, error)
	CreatePlayer(context.Context, string, string, interface{}) (string, error)
	DeleteMembership(context.Context, *DeleteMembershipPayload) (*Result, error)
	InviteForMembership(context.Context, *InvitationPayload) (*Result, error)
	LeaveClan(context.Context, string) (*LeaveClanResult, error)
	PromoteDemote(context.Context, *PromoteDemotePayload) (*Result, error)
	RetrieveClan(context.Context, string) (*Clan, error)
	RetrieveClansSummary(context.Context, []string) ([]*ClanSummary, error)
	RetrieveClanMembers(context.Context, string) (*ClanMembers, error)
	RetrieveClanSummary(context.Context, string) (*ClanSummary, error)
	RetrievePlayer(context.Context, string) (*Player, error)
	TransferOwnership(context.Context, string, string) (*TransferOwnershipResult, error)
	UpdateClan(context.Context, *ClanPayload) (*Result, error)
	UpdatePlayer(context.Context, string, string, interface{}) (*Result, error)
	SearchClans(context.Context, string) (*SearchClansResult, error)
	SearchClansWithOptions(context.Context, string, *SearchOptions) (*SearchClansResult, error)
}
