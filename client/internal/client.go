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

	ShadowsocksVPNConfig     string `bson:"shadowsocks_vpn_config" json:"shadowsocks_vpn_config"`
	VlessVPNConfig           string `bson:"vless_vpn_config" json:"vless_vpn_config"`
	CountryServerShadowsocks string `bson:"country_server_shadowsocks" json:"country_server_shadowsocks"`
	CountryServerVless       string `bson:"country_server_vless" json:"country_server_vless"`

	ConnectedSince time.Time `bson:"connected_since" json:"connected_since"`
	TimeLeft       time.Time `bson:"time_left" json:"time_left"`

	HasTrialBeenUsed bool `bson:"has_trial_been_used" json:"has_trial_been_used"`
	IsTrialActiveNow bool `bson:"is_trial_active_now" json:"is_trial_active_now"`
}
