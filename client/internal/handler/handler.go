package handler

import (
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
)

type Config struct{}

type Handler struct {
	cfg        Config
	clientRepo repository.ClientRepo
	pb.UnimplementedBulkaVPNServiceServer
}

func New(
	cfg Config,
	clientRepo repository.ClientRepo,
) pb.BulkaVPNServiceServer {
	return &Handler{
		cfg:        cfg,
		clientRepo: clientRepo,
	}

}
