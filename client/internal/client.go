package internal

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	TelegramID int64              `bson:"telegram_id,omitempty" json:"telegram_id"`

	ClientID string `bson:"client_id" json:"client_id"`
	Ver      int64  `bson:"ver" json:"ver"`

	OvpnConfig    string `bson:"ovpn_config" json:"ovpn_config"`
	CountryServer string `bson:"country_server" json:"country_server"`

	ConnectedSince time.Time `bson:"connected_since" json:"connected_since"`
	TimeLeft       time.Time `bson:"time_left" json:"time_left"`

	HasTrialBeenUsed bool `bson:"has_trial_been_used" json:"has_trial_been_used"`
	IsTrialActiveNow bool `bson:"is_trial_active_now" json:"is_trial_active_now"`
}
