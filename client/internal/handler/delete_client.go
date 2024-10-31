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
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{ClientID: req.ClientId})
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

	if keyDeleted {
		err = h.clientRepo.Delete(ctx, req.ClientId)
		if err != nil {
			return nil, fmt.Errorf("client.Delete: failed to delete client from database: %w", err)
		}
		return &pb.DeleteClientResponse{}, nil
	}

	return nil, fmt.Errorf("client.Delete: failed to delete key from VPN service or key was not deleted")
}
