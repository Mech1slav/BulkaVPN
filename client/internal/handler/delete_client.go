package handler

import (
	"context"
	"fmt"

	"BulkaVPN/client/countries/germany"
	"BulkaVPN/client/countries/holland"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
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

	switch client.CountryServer {
	case "Germany, Frankfurt":
		deleteKeyErr = germany.DeleteKeyByConfig(client.OvpnConfig)
		keyDeleted = germany.GetKey(client.OvpnConfig)

	case "Holland, Amsterdam":
		deleteKeyErr = holland.DeleteKeyByConfig(client.OvpnConfig)
		keyDeleted = holland.GetKey(client.OvpnConfig)

	default:
		return nil, fmt.Errorf("client.Delete: unsupported country server: %v", client.CountryServer)
	}

	if deleteKeyErr != nil {
		return nil, fmt.Errorf("client.Delete: failed to delete client from VPN service: %v", deleteKeyErr)
	}

	if req.IsTrialActiveNow {
		client.IsTrialActiveNow = false
	}

	client.OvpnConfig = ""
	client.Ver++

	if keyDeleted {
		err = h.clientRepo.Update(ctx, client, client.Ver)
		if err != nil {
			return nil, fmt.Errorf("client.Delete: failed to update client.OvpnConfig: %w", err)
		}
		return &pb.DeleteClientResponse{
			Deleted: keyDeleted,
		}, nil
	}

	return nil, fmt.Errorf("client.Delete: failed to delete key from VPN service or key was not deleted")
}
