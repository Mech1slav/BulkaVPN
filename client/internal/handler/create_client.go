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

func (h *Handler) CreateClient(ctx context.Context, req *pb.CreateClientRequest) (*pb.CreateClientResponse, error) {
	now := time.Now()
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client: %v", err)
	}

	// Если клиент не найден, возвращаем ошибку
	if client == nil {
		return nil, fmt.Errorf("client with Telegram ID %s does not exist", req.TelegramId)
	}

	// Проверка на истекшее или активное время доступа и продление/создание ключа
	if client.TimeLeft.After(now) {
		return &pb.CreateClientResponse{
			OvpnConfig: client.OvpnConfig,
			ClientId:   client.ClientID,
			TimeLeft:   timestamppb.New(client.TimeLeft),
		}, nil
	} else {
		client.TimeLeft = now.AddDate(0, 0, 30) // Устанавливаем на 30 дней
		var ovpnConfig string
		if req.CountryServer == "Holland, Amsterdam" {
			ovpnConfig, err = holland.CreateHollandVPNKey()
		} else if req.CountryServer == "Germany, Frankfurt" {
			ovpnConfig, err = germany.CreateGermanyVPNKey()
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create VPN key: %v", err)
		}

		client.OvpnConfig = ovpnConfig
		client.CountryServer = req.CountryServer

		if err := h.clientRepo.Update(ctx, client, client.Ver); err != nil {
			return nil, fmt.Errorf("failed to update client: %v", err)
		}

		return &pb.CreateClientResponse{
			OvpnConfig: client.OvpnConfig,
			ClientId:   client.ClientID,
			TimeLeft:   timestamppb.New(client.TimeLeft),
		}, nil
	}
}
