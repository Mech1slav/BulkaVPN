package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"BulkaVPN/client/countries/germany"
	"BulkaVPN/client/countries/holland"
	"BulkaVPN/client/internal"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
	"BulkaVPN/pkg/idstr"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handler) CreateTrialClient(ctx context.Context, req *pb.CreateTrialClientRequest) (*pb.CreateTrialClientResponse, error) {
	now := time.Now()

	// Сценарий при нажатии кнопки "/start"
	if req.StartButton {
		client := &internal.Client{
			ID:               primitive.NewObjectID(),
			TelegramID:       req.TelegramId,
			ClientID:         idstr.MustNew(8),
			Ver:              1,
			ConnectedSince:   now,
			HasTrialBeenUsed: false,
			IsTrialActiveNow: false,
		}
		if err := h.clientRepo.Create(ctx, client); err != nil {
			return nil, err
		}
		return &pb.CreateTrialClientResponse{}, nil
	}

	// Сценарий при нажатии "Получить тестовый ключ"
	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
	if err != nil || client == nil {
		return nil, errors.New("client not found")
	}

	// Проверка статуса клиента
	if !req.Trial {
		switch {
		case !client.HasTrialBeenUsed && !client.IsTrialActiveNow:
			return &pb.CreateTrialClientResponse{
				CountryServer: "Вы можете выбрать локацию для пробного периода",
			}, nil
		case client.HasTrialBeenUsed && !client.IsTrialActiveNow:
			return nil, errors.New("пробный период уже был использован")
		case client.IsTrialActiveNow:
			return &pb.CreateTrialClientResponse{
				OvpnConfig:    client.OvpnConfig,
				CountryServer: client.CountryServer,
				TimeLeft:      timestamppb.New(client.TimeLeft),
			}, nil
		}
	}

	// Сценарий активации пробного ключа
	if req.Trial {
		var ovpnConfig string
		switch req.CountryServer {
		case "Holland, Amsterdam":
			ovpnConfig, err = holland.CreateHollandVPNKey()
		case "Germany, Frankfurt":
			ovpnConfig, err = germany.CreateGermanyVPNKey()
		default:
			return nil, fmt.Errorf("unknown country server: %v", req.CountryServer)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create VPN key: %v", err)
		}

		client.OvpnConfig = ovpnConfig
		client.CountryServer = req.CountryServer
		client.TimeLeft = now.Add(72 * time.Hour) // Установить срок в 3 дня
		client.HasTrialBeenUsed = true
		client.IsTrialActiveNow = true
		client.Ver++

		if err := h.clientRepo.Update(ctx, client, client.Ver); err != nil {
			return nil, fmt.Errorf("failed to update client: %v", err)
		}

		return &pb.CreateTrialClientResponse{
			OvpnConfig:    ovpnConfig,
			CountryServer: req.CountryServer,
			TimeLeft:      timestamppb.New(client.TimeLeft),
		}, nil
	}

	return nil, errors.New("invalid request parameters")
}
