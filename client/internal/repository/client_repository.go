package repository

import (
	"context"

	"BulkaVPN/client/internal"
	pb "BulkaVPN/client/proto"
)

type ClientRepo interface {
	Create(ctx context.Context, client *internal.Client) error
	Get(ctx context.Context, opts ClientGetOpts) (*internal.Client, error)
	Search(ctx context.Context, opts ClientSearchOpts) ([]*internal.Client, error)
	Update(ctx context.Context, client *internal.Client, versionCheck int64) error
	Delete(ctx context.Context, clientID string) error
	Count(ctx context.Context, opts ClientSearchOpts) (int64, error)
}

type ClientSearchOpts struct {
	Filter  *pb.ClientFilter
	Limit   int64
	Skip    int64
	AfterID string
}

type ClientGetOpts struct {
	ClientID             string
	ShadowsocksVPNConfig string
	VlessVPNConfig       string
	TelegramID           int64
}
