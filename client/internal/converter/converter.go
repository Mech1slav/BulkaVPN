package converter

import (
	"BulkaVPN/client/internal"
	pb "BulkaVPN/client/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Clients(i *internal.Client) *pb.Client {
	res := &pb.Client{
		ClientId:         i.ClientID,
		TelegramId:       i.TelegramID,
		Ver:              i.Ver,
		OvpnConfig:       i.OvpnConfig,
		CountryServer:    i.CountryServer,
		HasTrialBeenUsed: i.HasTrialBeenUsed,
		IsTrialActiveNow: i.IsTrialActiveNow,
		ConnectedSince:   timestamppb.New(i.ConnectedSince),
		TimeLeft:         timestamppb.New(i.TimeLeft),
	}

	return res
}
