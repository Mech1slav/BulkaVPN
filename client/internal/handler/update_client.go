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

func (h *Handler) UpdateClient(ctx context.Context, req *pb.UpdateClientRequest) (*pb.UpdateClientResponse, error) {
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{
		TelegramID: req.TelegramId,
	})
	if err != nil {
		return nil, err
	}

	var (
		newShadowsocksVPNConfig string
		newVlessVPNConfig       string
	)
	switch req.CountryServerShadowsocks {
	case "Holland, Amsterdam":
		newShadowsocksVPNConfig, err = ssholland.CreateHollandVPNKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create Holland VPN key: %v", err)
		}
		if client.CountryServerShadowsocks == "Germany, Frankfurt" {
			if err := ssgermany.DeleteKeyByConfig(client.ShadowsocksVPNConfig); err != nil {
				return nil, fmt.Errorf("client.Delete: failed to delete client from germany_shadowsocks vpn service: %v", err)
			}
		}
	case "Germany, Frankfurt":
		newShadowsocksVPNConfig, err = ssgermany.CreateGermanyVPNKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create Germany VPN key: %v", err)
		}
		if client.CountryServerShadowsocks == "Holland, Amsterdam" {
			if err := ssholland.DeleteKeyByConfig(client.ShadowsocksVPNConfig); err != nil {
				return nil, fmt.Errorf("client.Delete: failed to delete client from holland_shadowsocks vpn service: %v", err)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported country: %s", req.CountryServerShadowsocks)
	}

	switch req.CountryServerVless {
	case "Holland, Amsterdam":
		newVlessVPNConfig, err = vlessholland.GenerateVPNKey(strconv.FormatInt(req.TelegramId, 10))
		if err != nil {
			return nil, fmt.Errorf("failed to create Holland VPN key: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported country: %s", req.CountryServerShadowsocks)
	}

	client.ShadowsocksVPNConfig = newShadowsocksVPNConfig
	client.CountryServerShadowsocks = req.CountryServerShadowsocks
	client.VlessVPNConfig = newVlessVPNConfig
	client.CountryServerVless = req.CountryServerVless
	client.Ver++

	if err := h.clientRepo.Update(ctx, client, client.Ver); err != nil {
		return nil, fmt.Errorf("client.Update: failed to update client: %v", err)
	}

	return &pb.UpdateClientResponse{
		ShadowsocksVpnConfig:     client.ShadowsocksVPNConfig,
		VlessVpnConfig:           client.VlessVPNConfig,
		CountryServerShadowsocks: req.CountryServerShadowsocks,
		CountryServerVless:       req.CountryServerVless,
	}, nil
}
