package handler

import (
	"context"

	"BulkaVPN/client/internal/converter"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
)

func (h *Handler) SearchClients(ctx context.Context, req *pb.SearchClientsRequest) (*pb.SearchClientsResponse, error) {
	var reqClientOpts *pb.ClientFilter

	if req.Filter == nil {
		reqClientOpts = &pb.ClientFilter{}
	} else {
		reqClientOpts = &pb.ClientFilter{
			ClientId:      req.Filter.ClientId,
			OvpnConfig:    req.Filter.OvpnConfig,
			CountryServer: req.Filter.CountryServer,
		}
	}

	clientOpts := repository.ClientSearchOpts{
		Filter: reqClientOpts,
	}

	var err error

	clients, err := h.clientRepo.Search(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	res := &pb.SearchClientsResponse{Clients: make([]*pb.Client, 0, len(clients))}

	for _, s := range clients {
		res.Clients = append(res.Clients, converter.Clients(s))
	}

	return res, err
}
