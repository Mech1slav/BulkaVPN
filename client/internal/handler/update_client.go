package handler

import (
	"context"
	"fmt"

	"BulkaVPN/client/countries/germany"
	"BulkaVPN/client/countries/holland"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
)

func (h *Handler) UpdateClient(ctx context.Context, req *pb.UpdateClientRequest) (*pb.UpdateClientResponse, error) {
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{
		TelegramID: req.TelegramId,
	})
	if err != nil {
		return nil, err
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
	client.Ver++

	if err := h.clientRepo.Update(ctx, client, client.Ver); err != nil {
		return nil, fmt.Errorf("client.Update: failed to update client: %v", err)
	}

	return &pb.UpdateClientResponse{
		CountryServer: req.CountryServer,
		OvpnConfig:    client.OvpnConfig,
	}, nil
}
