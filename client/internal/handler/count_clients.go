package handler

import (
	"context"

	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
)

func (h *Handler) CountClients(ctx context.Context, req *pb.CountClientsRequest) (*pb.CountClientsResponse, error) {
	var reqClientOpts *pb.ClientFilter

	if req.Filter == nil {
		reqClientOpts = &pb.ClientFilter{}
	} else {
		reqClientOpts = &pb.ClientFilter{
			OvpnConfig:    req.Filter.OvpnConfig,
			CountryServer: req.Filter.CountryServer,
		}
	}

	clientOpts := repository.ClientSearchOpts{
		Filter: reqClientOpts,
	}

	client, err := h.clientRepo.Count(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	return &pb.CountClientsResponse{Count: int32(client)}, nil
}
