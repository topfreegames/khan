package lib

import "context"

// KhanInterface defines the interface for the khan client
type KhanInterface interface {
	CreatePlayer(context.Context, string, string, interface{}) (string, error)
	UpdatePlayer(context.Context, string, string, interface{}) error
	RetrievePlayer(context.Context, string) (*Player, error)
	CreateClan(context.Context, *ClanPayload) (string, error)
	UpdateClan(context.Context, *ClanPayload) error
	RetrieveClanSummary(context.Context, string) (*ClanSummary, error)
	RetrieveClansSummary(context.Context, []string) ([]*ClanSummary, error)
}
