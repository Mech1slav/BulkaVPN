package handler

import (
	"context"
	"fmt"
	"strconv"

	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
	ssgermany "BulkaVPN/client/protocols/shadowsocks/germany_shadowsocks"
	ssholland "BulkaVPN/client/protocols/shadowsocks/holland_shadowsocks"
	vlessholland "BulkaVPN/client/protocols/vless/holland_vless"
)

func (h *Handler) DeleteClient(ctx context.Context, req *pb.DeleteClientRequest) (*pb.DeleteClientResponse, error) {
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{
		TelegramID: req.TelegramId,
	})
	if err != nil {
		return nil, fmt.Errorf("client.Delete: failed to find client in database: %v", err)
	}

	var deleteKeyErr error
	var keyDeleted bool

	switch client.CountryServerShadowsocks {
	case "Germany, Frankfurt":
		deleteKeyErr = ssgermany.DeleteKeyByConfig(client.ShadowsocksVPNConfig)
		keyDeleted = ssgermany.GetKey(client.ShadowsocksVPNConfig)

	case "Holland, Amsterdam":
		deleteKeyErr = ssholland.DeleteKeyByConfig(client.ShadowsocksVPNConfig)
		keyDeleted = ssholland.GetKey(client.ShadowsocksVPNConfig)

	default:
		return nil, fmt.Errorf("client.Delete: unsupported country server: %v", client.CountryServerShadowsocks)
	}

	switch client.CountryServerVless {
	case "Holland, Amsterdam":
		deleteKeyErr = vlessholland.DeleteKeyByConfig(strconv.FormatInt(client.TelegramID, 10))
		keyDeleted = vlessholland.GetKeyByConfig(strconv.FormatInt(client.TelegramID, 10))
	}

	if deleteKeyErr != nil {
		return nil, fmt.Errorf("client.Delete: failed to delete client from VPN service: %v", deleteKeyErr)
	}

	if req.IsTrialActiveNow {
		client.IsTrialActiveNow = false
	}

	client.ShadowsocksVPNConfig = ""
	client.VlessVPNConfig = ""
	client.Ver++

	if keyDeleted {
		err = h.clientRepo.Update(ctx, client, client.Ver)
		if err != nil {
			return nil, fmt.Errorf("client.Delete: failed to update client.ShadowsocksVPNConfig: %w", err)
		}
		return &pb.DeleteClientResponse{
			Deleted: keyDeleted,
		}, nil
	}

	return nil, fmt.Errorf("client.Delete: failed to delete key from VPN service or key was not deleted")
}
