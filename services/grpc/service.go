package grpc

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/cpchain-network/gas-oracle/database"
	"github.com/cpchain-network/gas-oracle/proto/gasfee"
)

const MaxRecvMessageSize = 1024 * 1024 * 30000

type GasFeeRpcConfig struct {
	Host string
	Port int
}

type GasFeeRpcService struct {
	*GasFeeRpcConfig

	db *database.DB

	gasfee.UnimplementedGasFeeServicesServer
	stopped atomic.Bool
}

func NewGasFeeRpcService(conf *GasFeeRpcConfig, db *database.DB) (*GasFeeRpcService, error) {
	return &GasFeeRpcService{
		GasFeeRpcConfig: conf,
		db:              db,
	}, nil
}

func (ms *GasFeeRpcService) Start(ctx context.Context) error {
	go func(ms *GasFeeRpcService) {
		rpcAddr := fmt.Sprintf("%s:%d", ms.GasFeeRpcConfig.Host, ms.GasFeeRpcConfig.Port)
		listener, err := net.Listen("tcp", rpcAddr)
		if err != nil {
			log.Error("Could not start tcp listener. ")
		}

		opt := grpc.MaxRecvMsgSize(MaxRecvMessageSize)

		gs := grpc.NewServer(
			opt,
			grpc.ChainUnaryInterceptor(
				nil,
			),
		)

		reflection.Register(gs)
		gasfee.RegisterGasFeeServicesServer(gs, ms)

		log.Info("grpc info", "addr", listener.Addr())

		if err := gs.Serve(listener); err != nil {
			log.Error("start rpc server fail", "err", err)
		}
	}(ms)
	return nil
}

func (ms *GasFeeRpcService) Stop(ctx context.Context) error {
	ms.stopped.Store(true)
	return nil
}

func (ms *GasFeeRpcService) Stopped() bool {
	return ms.stopped.Load()
}
