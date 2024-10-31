package repository

import (
	"context"
	"time"

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
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	SaveUser(ctx context.Context, user *User) error
}

type ClientSearchOpts struct {
	Filter  *pb.ClientFilter
	Limit   int64
	Skip    int64
	AfterID string
}

type ClientGetOpts struct {
	ClientID   string
	OvpnConfig string
}

type User struct {
	TelegramID int64     `bson:"telegram_id"`
	HasTrial   bool      `bson:"has_trial"`
	TrialEnd   time.Time `bson:"trial_end,omitempty"`
}
