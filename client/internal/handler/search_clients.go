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
			ClientId:                 req.Filter.ClientId,
			ShadowsocksVpnConfig:     req.Filter.ShadowsocksVpnConfig,
			CountryServerShadowsocks: req.Filter.CountryServerShadowsocks,
			VlessVpnConfig:           req.Filter.VlessVpnConfig,
			CountryServerVless:       req.Filter.CountryServerVless,
			TelegramId:               req.Filter.TelegramId,
			HasTrialBeenUsed:         req.Filter.HasTrialBeenUsed,
			IsTrialActiveNow:         req.Filter.IsTrialActiveNow,
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
