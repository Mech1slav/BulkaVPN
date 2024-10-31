package handler

import (
	"context"
	"fmt"
	"time"

	"BulkaVPN/client/countries/germany"
	"BulkaVPN/client/countries/holland"
	"BulkaVPN/client/internal"
	pb "BulkaVPN/client/proto"
	"BulkaVPN/pkg/idstr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handler) CreateClient(ctx context.Context, req *pb.CreateClientRequest) (*pb.CreateClientResponse, error) {
	clientID := idstr.MustNew(8)
	now := time.Now()
	timeLeft := now.AddDate(0, 0, 30)
	timeLeftTrial := now.AddDate(0, 0, 3)
	selectedCountry := req.CountryServer

	if req.CountryServer == "Holland, Amsterdam" && req.Trial == "false" {
		selectedCountry = req.CountryServer

		ovpnConfig, err := holland.CreateHollandVPNKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create VPN key: %v", err)
		}

		client := &internal.Client{
			ClientID:       clientID,
			Ver:            1,
			ConnectedSince: now,
			OvpnConfig:     ovpnConfig,
			CountryServer:  selectedCountry,
			TimeLeft:       timeLeft,
		}

		if err := h.clientRepo.Create(ctx, client); err != nil {
			return nil, err
		}

		return &pb.CreateClientResponse{
			OvpnConfig: ovpnConfig,
			ClientId:   clientID,
			TimeLeft:   timestamppb.New(timeLeft),
		}, nil

	}

	if req.CountryServer == "Germany, Frankfurt" && req.Trial == "false" {
		selectedCountry = req.CountryServer

		ovpnConfig, err := germany.CreateGermanyVPNKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create VPN key: %v", err)
		}

		client := &internal.Client{
			ClientID:       clientID,
			Ver:            1,
			ConnectedSince: now,
			OvpnConfig:     ovpnConfig,
			CountryServer:  selectedCountry,
			TimeLeft:       timeLeft,
		}

		if err := h.clientRepo.Create(ctx, client); err != nil {
			return nil, err
		}

		return &pb.CreateClientResponse{
			OvpnConfig: ovpnConfig,
			ClientId:   clientID,
			TimeLeft:   timestamppb.New(timeLeft),
		}, nil
	}

	if req.CountryServer == "Holland, Amsterdam" && req.Trial == "true" {
		selectedCountry = req.CountryServer

		ovpnConfig, err := holland.CreateHollandVPNKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create VPN key: %v", err)
		}

		client := &internal.Client{
			ClientID:       clientID,
			Ver:            1,
			ConnectedSince: now,
			OvpnConfig:     ovpnConfig,
			CountryServer:  selectedCountry,
			TimeLeft:       timeLeftTrial,
		}

		if err := h.clientRepo.Create(ctx, client); err != nil {
			return nil, err
		}

		return &pb.CreateClientResponse{
			OvpnConfig: ovpnConfig,
			ClientId:   clientID,
			TimeLeft:   timestamppb.New(timeLeftTrial),
		}, nil

	}

	return nil, fmt.Errorf("unsupported country: %s", req.CountryServer)
}
