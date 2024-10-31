package main

import (
	"context"

	"BulkaVPN/client/internal/handler"
	repomongo "BulkaVPN/client/internal/repository/mongo"
	pb "BulkaVPN/client/proto"
	"BulkaVPN/pkg/config"
	"BulkaVPN/pkg/grpcx"
	"BulkaVPN/pkg/logx"
	"BulkaVPN/pkg/logx/zap"
	"BulkaVPN/pkg/mongox"
	"BulkaVPN/pkg/natsx"
	"BulkaVPN/pkg/signalx"
	"BulkaVPN/pkg/tracing"
	"BulkaVPN/pkg/tracing/jaeger"
)

type Config struct {
	Zap        zap.Config
	Jaeger     jaeger.Config
	Mongo      mongox.Config
	GRPCServer grpcx.ServerConfig
	Nats       natsx.Config

	ClientRepository repomongo.ClientConfig

	Handler handler.Config

	ClientsAddr string `envconfig:"CLIENTS_ADDR" default:"localhost:50051"`
}

func main() {
	if err := run(context.Background()); err != nil {
		logx.NewSimple().Errorf(err.Error())
	}
}

func run(ctx context.Context) error {
	var cfg Config

	if err := config.Process(&cfg); err != nil {
		return err
	}

	l, err := zap.New(cfg.Zap)
	if err != nil {
		return err
	}

	ctx = l.ToCtx(ctx)

	tracer, clean, err := jaeger.New(cfg.Jaeger)
	if err != nil {
		return err
	}

	ctx = tracing.ToCtx(ctx, tracer)

	defer clean()

	mongoDB, clean, err := mongox.New(ctx, cfg.Mongo)
	if err != nil {
		return err
	}

	defer clean()

	clientRepo := repomongo.NewClientRepo(cfg.ClientRepository, mongoDB)

	h := handler.New(cfg.Handler, clientRepo)

	srv, err := grpcx.New(ctx, cfg.GRPCServer)
	if err != nil {
		return err
	}

	pb.RegisterBulkaVPNServiceServer(srv, h)

	if err = srv.Start(ctx, cfg.ClientsAddr); err != nil {
		return err
	}

	defer srv.Stop()

	signalx.Wait()

	return nil
}
