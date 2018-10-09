package lib

import "context"

// KhanInterface defines the interface for the khan client
type KhanInterface interface {
	ApplyForMembership(context.Context, *ApplicationPayload) (*ClanApplyResult, error)
	ApproveDenyMembershipApplication(context.Context, *ApplicationApprovalPayload) error
	ApproveDenyMembershipInvitation(context.Context, *InvitationApprovalPayload) error
	CreateClan(context.Context, *ClanPayload) (string, error)
	CreatePlayer(context.Context, string, string, interface{}) (string, error)
	DeleteMembership(context.Context, *DeleteMembershipPayload) error
	InviteForMembership(context.Context, *InvitationPayload) error
	LeaveClan(context.Context, string) (*LeaveClanResult, error)
	PromoteDemote(context.Context, *PromoteDemotePayload) error
	RetrieveClan(context.Context, string) (*Clan, error)
	RetrieveClansSummary(context.Context, []string) ([]*ClanSummary, error)
	RetrieveClanSummary(context.Context, string) (*ClanSummary, error)
	RetrievePlayer(context.Context, string) (*Player, error)
	TransferOwnership(context.Context, string, string) (*TransferOwnershipResult, error)
	UpdateClan(context.Context, *ClanPayload) error
	UpdatePlayer(context.Context, string, string, interface{}) error
}
