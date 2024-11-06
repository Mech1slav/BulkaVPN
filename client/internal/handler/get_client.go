package handler

import (
	"context"

	"BulkaVPN/client/internal/converter"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
)

func (h *Handler) GetClient(ctx context.Context, req *pb.GetClientRequest) (*pb.GetClientResponse, error) {
	var err error

	opts := repository.ClientGetOpts{
		ClientID:   req.ClientId,
		OvpnConfig: req.OvpnConfig,
		TelegramID: req.TelegramId,
	}

	res, err := h.clientRepo.Get(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &pb.GetClientResponse{Clients: converter.Clients(res)}, nil
}
