package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
	"BulkaVPN/client/protocols/shadowsocks/germany_shadowsocks"
	"BulkaVPN/client/protocols/shadowsocks/holland_shadowsocks"
)

func (h *Handler) CreateClient(ctx context.Context, req *pb.CreateClientRequest) (*pb.CreateClientResponse, error) {
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
	if err != nil || client == nil {
		return nil, errors.New("client not found")
	}

	var ovpnConfig string
	switch req.CountryServer {
	case "Holland, Amsterdam":
		ovpnConfig, err = holland_shadowsocks.CreateHollandVPNKey()
	case "Germany, Frankfurt":
		ovpnConfig, err = germany_shadowsocks.CreateGermanyVPNKey()
	default:
		return nil, fmt.Errorf("unknown server: %v", req.CountryServer)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create new vpn key: %v", err)
	}

	now := time.Now()
	newTimeLeft := req.TimeLeft.AsTime()
	if client.TimeLeft.Before(now) {
		client.TimeLeft = newTimeLeft
	} else {
		client.TimeLeft = client.TimeLeft.Add(newTimeLeft.Sub(now))
	}

	client.OvpnConfig = ovpnConfig
	client.CountryServer = req.CountryServer
	client.Ver++

	if err := h.clientRepo.Update(ctx, client, client.Ver); err != nil {
		return nil, fmt.Errorf("failed to update client data: %v", err)
	}

	return &pb.CreateClientResponse{
		OvpnConfig:    client.OvpnConfig,
		CountryServer: client.CountryServer,
		TimeLeft:      timestamppb.New(client.TimeLeft),
	}, nil
}
