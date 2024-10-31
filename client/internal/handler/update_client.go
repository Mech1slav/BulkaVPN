package handler

import (
	"context"
	"fmt"
	"time"

	"BulkaVPN/client/countries/germany"
	"BulkaVPN/client/countries/holland"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handler) UpdateClient(ctx context.Context, req *pb.UpdateClientRequest) (*pb.UpdateClientResponse, error) {
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{ClientID: req.ClientId})
	if err != nil {
		return nil, err
	}

	if req.CountryServer == client.CountryServer {
		return nil, fmt.Errorf("client.Update: You are already connected to this server")
	}

	var newOvpnConfig string
	switch req.CountryServer {
	case "Holland, Amsterdam":
		newOvpnConfig, err = holland.CreateHollandVPNKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create Holland VPN key: %v", err)
		}
		if client.CountryServer == "Germany, Frankfurt" {
			if err := germany.DeleteKeyByConfig(client.OvpnConfig); err != nil {
				return nil, fmt.Errorf("client.Delete: failed to delete client from germany vpn service: %v", err)
			}
		}
	case "Germany, Frankfurt":
		newOvpnConfig, err = germany.CreateGermanyVPNKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create Germany VPN key: %v", err)
		}
		if client.CountryServer == "Holland, Amsterdam" {
			if err := holland.DeleteKeyByConfig(client.OvpnConfig); err != nil {
				return nil, fmt.Errorf("client.Delete: failed to delete client from holland vpn service: %v", err)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported country: %s", req.CountryServer)
	}

	client.OvpnConfig = newOvpnConfig
	client.CountryServer = req.CountryServer
	client.Ver += 1

	if err := h.clientRepo.Update(ctx, client, client.Ver); err != nil {
		return nil, fmt.Errorf("client.Update: failed to update client: %v", err)
	}

	now := time.Now()
	timeLeft := client.ConnectedSince.Add(30 * 24 * time.Hour).Sub(now)

	return &pb.UpdateClientResponse{
		CountryServer: req.CountryServer,
		TimeLeft:      timestamppb.New(now.Add(timeLeft)), // Вернем оставшееся время
	}, nil
}
