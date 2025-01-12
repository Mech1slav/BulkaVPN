package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	"BulkaVPN/client/internal"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
	"BulkaVPN/client/protocols/shadowsocks/germany_shadowsocks"
	"BulkaVPN/client/protocols/shadowsocks/holland_shadowsocks"
	"BulkaVPN/client/protocols/vless/holland_vless"
	"BulkaVPN/pkg/idstr"
)

func (h *Handler) CreateTrialClient(ctx context.Context, req *pb.CreateTrialClientRequest) (*pb.CreateTrialClientResponse, error) {
	now := time.Now()

	if req.StartButton {
		getClient, err := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
		if err != nil && getClient == nil {
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
				CountryServerShadowsocks: "Вы можете выбрать локацию для тестового периода",
			}, nil
		case client.HasTrialBeenUsed && !client.IsTrialActiveNow:
			return &pb.CreateTrialClientResponse{
				CountryServerShadowsocks: "Тестовый период уже был использован",
			}, nil
		case client.IsTrialActiveNow:
			return &pb.CreateTrialClientResponse{
				ShadowsocksVpnConfig:     client.ShadowsocksVPNConfig,
				VlessVpnConfig:           client.VlessVPNConfig,
				CountryServerShadowsocks: client.CountryServerShadowsocks,
				CountryServerVless:       client.CountryServerVless,
				TimeLeft:                 timestamppb.New(client.TimeLeft),
			}, nil
		}
	}

	if req.Trial {
		newClient, err2 := h.clientRepo.Get(ctx, repository.ClientGetOpts{TelegramID: req.TelegramId})
		if newClient.HasTrialBeenUsed == false && newClient.IsTrialActiveNow == false {
			var (
				shadowsocksVPNConfig string
				vlessVPNConfig       string
			)
			switch req.CountryServerShadowsocks {
			case "Holland, Amsterdam":
				shadowsocksVPNConfig, err2 = holland_shadowsocks.CreateHollandVPNKey()
			case "Germany, Frankfurt":
				shadowsocksVPNConfig, err2 = germany_shadowsocks.CreateGermanyVPNKey()
			default:
				return nil, fmt.Errorf("unknown country server: %v", req.CountryServerShadowsocks)
			}

			switch req.CountryServerVless {
			case "Holland, Amsterdam":
				vlessVPNConfig, err2 = holland_vless.GenerateVPNKey(strconv.FormatInt(req.TelegramId, 10))
			}

			if err2 != nil {
				return nil, fmt.Errorf("failed to create VPN key: %v", err2)
			}

			newClient.ShadowsocksVPNConfig = shadowsocksVPNConfig
			newClient.CountryServerShadowsocks = req.CountryServerShadowsocks
			newClient.VlessVPNConfig = vlessVPNConfig
			newClient.CountryServerVless = req.CountryServerVless
			newClient.TimeLeft = now.Add(72 * time.Hour)
			newClient.HasTrialBeenUsed = true
			newClient.IsTrialActiveNow = true
			newClient.Ver++

			if err := h.clientRepo.Update(ctx, newClient, newClient.Ver); err != nil {
				return nil, fmt.Errorf("failed to update newClient: %v", err)
			}

			return &pb.CreateTrialClientResponse{
				ShadowsocksVpnConfig:     shadowsocksVPNConfig,
				CountryServerShadowsocks: req.CountryServerShadowsocks,
				VlessVpnConfig:           vlessVPNConfig,
				CountryServerVless:       req.CountryServerVless,
				TimeLeft:                 timestamppb.New(newClient.TimeLeft),
			}, nil
		}
	}

	return nil, errors.New("invalid request parameters")
}
