package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
	"BulkaVPN/client/protocols/shadowsocks/germany_shadowsocks"
	"BulkaVPN/client/protocols/shadowsocks/holland_shadowsocks"
	"BulkaVPN/client/protocols/vless/holland_vless"
)

func (h *Handler) CreateClient(ctx context.Context, req *pb.CreateClientRequest) (*pb.CreateClientResponse, error) {
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
	if err != nil || client == nil {
		return nil, errors.New("client not found")
	}

	var (
		shadowsocksVPNConfig string
		vlessVPNConfig       string
	)
	switch req.CountryServerShadowsocks {
	case "Holland, Amsterdam":
		shadowsocksVPNConfig, err = holland_shadowsocks.CreateHollandVPNKey()
	case "Germany, Frankfurt":
		shadowsocksVPNConfig, err = germany_shadowsocks.CreateGermanyVPNKey()
	default:
		return nil, fmt.Errorf("unknown server: %v", req.CountryServerShadowsocks)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create new vpn key: %v", err)
	}

	switch req.CountryServerVless {
	case "Holland, Amsterdam":
		vlessVPNConfig, err = holland_vless.GenerateVPNKey(strconv.FormatInt(req.TelegramId, 10))
	default:
		return nil, fmt.Errorf("unknown server: %v", req.CountryServerVless)
	}

	now := time.Now()
	newTimeLeft := req.TimeLeft.AsTime()
	if client.TimeLeft.Before(now) {
		client.TimeLeft = newTimeLeft
	} else {
		client.TimeLeft = client.TimeLeft.Add(newTimeLeft.Sub(now))
	}

	client.ShadowsocksVPNConfig = shadowsocksVPNConfig
	client.VlessVPNConfig = vlessVPNConfig
	client.CountryServerShadowsocks = req.CountryServerShadowsocks
	client.CountryServerVless = req.CountryServerVless
	client.Ver++

	if err := h.clientRepo.Update(ctx, client, client.Ver); err != nil {
		return nil, fmt.Errorf("failed to update client data: %v", err)
	}

	return &pb.CreateClientResponse{
		ShadowsocksVpnConfig:     client.ShadowsocksVPNConfig,
		VlessVpnConfig:           client.VlessVPNConfig,
		CountryServerShadowsocks: client.CountryServerShadowsocks,
		CountryServerVless:       client.CountryServerVless,
		TimeLeft:                 timestamppb.New(client.TimeLeft),
	}, nil
}
