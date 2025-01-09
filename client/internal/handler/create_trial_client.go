package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	"BulkaVPN/client/internal"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
	"BulkaVPN/client/protocols/shadowsocks/germany_shadowsocks"
	"BulkaVPN/client/protocols/shadowsocks/holland_shadowsocks"
	"BulkaVPN/pkg/idstr"
)

func (h *Handler) CreateTrialClient(ctx context.Context, req *pb.CreateTrialClientRequest) (*pb.CreateTrialClientResponse, error) {
	now := time.Now()

	if req.StartButton {
		client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
		if err != nil && client == nil {
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
	}

	client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
	if err != nil || client == nil {
		return nil, errors.New("client not found")
	}

	if !req.Trial {
		switch {
		case !client.HasTrialBeenUsed && !client.IsTrialActiveNow:
			return &pb.CreateTrialClientResponse{
				CountryServer: "Вы можете выбрать локацию для тестового периода",
			}, nil
		case client.HasTrialBeenUsed && !client.IsTrialActiveNow:
			return &pb.CreateTrialClientResponse{
				CountryServer: "Тестовый период уже был использован",
			}, nil
		case client.IsTrialActiveNow:
			return &pb.CreateTrialClientResponse{
				OvpnConfig:    client.OvpnConfig,
				CountryServer: client.CountryServer,
				TimeLeft:      timestamppb.New(client.TimeLeft),
			}, nil
		}
	}

	if req.Trial {
		client, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
		if client.HasTrialBeenUsed == false && client.IsTrialActiveNow == false {
			var ovpnConfig string
			switch req.CountryServer {
			case "Holland, Amsterdam":
				ovpnConfig, err = holland_shadowsocks.CreateHollandVPNKey()
			case "Germany, Frankfurt":
				ovpnConfig, err = germany_shadowsocks.CreateGermanyVPNKey()
			default:
				return nil, fmt.Errorf("unknown country server: %v", req.CountryServer)
			}

			if err != nil {
				return nil, fmt.Errorf("failed to create VPN key: %v", err)
			}

			client.OvpnConfig = ovpnConfig
			client.CountryServer = req.CountryServer
			client.TimeLeft = now.Add(72 * time.Hour)
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
	}

	return nil, errors.New("invalid request parameters")
}
