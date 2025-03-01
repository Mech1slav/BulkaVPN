package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"BulkaVPN/client/internal"
	pb "BulkaVPN/client/proto"
)

func Clients(i *internal.Client) *pb.Client {
	res := &pb.Client{
		ClientId:                 i.ClientID,
		TelegramId:               i.TelegramID,
		Ver:                      i.Ver,
		ShadowsocksVpnConfig:     i.ShadowsocksVPNConfig,
		VlessVpnConfig:           i.VlessVPNConfig,
		CountryServerShadowsocks: i.CountryServerShadowsocks,
		CountryServerVless:       i.CountryServerVless,
		HasTrialBeenUsed:         i.HasTrialBeenUsed,
		IsTrialActiveNow:         i.IsTrialActiveNow,
		ConnectedSince:           timestamppb.New(i.ConnectedSince),
		TimeLeft:                 timestamppb.New(i.TimeLeft),
	}

	return res
}
