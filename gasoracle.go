package gas_oracle

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/cpchain-network/gas-oracle/config"
	"github.com/cpchain-network/gas-oracle/database"
	"github.com/cpchain-network/gas-oracle/synchronizer"
	"github.com/cpchain-network/gas-oracle/synchronizer/node"
)

type GasOracle struct {
	db           *database.DB
	ethClient    map[uint64]node.EthClient
	Synchronizer map[uint64]*synchronizer.OracleSynchronizer
	shutdown     context.CancelCauseFunc
	stopped      atomic.Bool
	backOffset   uint64
	loopInternal time.Duration
	chainIdList  []uint64
}

func NewGasOracle(ctx context.Context, cfg *config.Config, shutdown context.CancelCauseFunc) (*GasOracle, error) {
	log.Info("new gas oracle start️ 🕖")
	out := &GasOracle{
		loopInternal: cfg.LoopInternal,
		backOffset:   cfg.BackOffset,
		shutdown:     shutdown,
	}
	if err := out.initFromConfig(ctx, cfg); err != nil {
		return nil, errors.Join(err, out.Stop(ctx))
	}
	log.Info("new gas oracle success🏅️")
	return out, nil
}

func (as *GasOracle) Start(ctx context.Context) error {
	for i := range as.chainIdList {
		log.Info("starting sync", "chainId", as.chainIdList[i])
		realChainId := as.chainIdList[i]
		if err := as.Synchronizer[realChainId].Start(context.Background()); err != nil {
			return fmt.Errorf("failed to start chain sync: %w", err)
		}
	}
	return nil
}

func (as *GasOracle) Stop(ctx context.Context) error {
	var result error
	for i := range as.chainIdList {
		if as.Synchronizer[as.chainIdList[i]] != nil {
			if err := as.Synchronizer[as.chainIdList[i]].Stop(ctx); err != nil {
				result = errors.Join(result, fmt.Errorf("failed to close synchronizer: %w", err))
			}
		}
		if as.ethClient[as.chainIdList[i]] != nil {
			as.ethClient[as.chainIdList[i]].Close()
		}
	}

	if as.db != nil {
		if err := as.db.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close DB: %w", err))
		}
	}

	as.stopped.Store(true)

	log.Info("gas oracle stopped")

	return result
}

func (as *GasOracle) Stopped() bool {
	return as.stopped.Load()
}

func (as *GasOracle) initFromConfig(ctx context.Context, cfg *config.Config) error {
	if err := as.initRPCClients(ctx, cfg); err != nil {
		return fmt.Errorf("failed to start RPC clients: %w", err)
	}

	if err := as.initDB(ctx, cfg.MasterDb); err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}

	if err := as.initSynchronizer(cfg); err != nil {
		return fmt.Errorf("failed to init L1 Sync: %w", err)
	}

	return nil
}

func (as *GasOracle) initRPCClients(ctx context.Context, conf *config.Config) error {
	for i := range conf.RPCs {
		log.Info("Init rpc client", "ChainId", conf.RPCs[i].ChainId, "RpcUrl", conf.RPCs[i].RpcUrl)
		rpc := conf.RPCs[i]
		ethClient, err := node.DialEthClient(ctx, rpc.RpcUrl)
		if err != nil {
			log.Error("dial eth client fail", "err", err)
			return fmt.Errorf("failed to dial L1 client: %w", err)
		}
		if as.ethClient == nil {
			as.ethClient = make(map[uint64]node.EthClient)
		}
		as.ethClient[rpc.ChainId] = ethClient
		as.chainIdList = append(as.chainIdList, rpc.ChainId)
	}
	log.Info("Init rpc client success")
	return nil
}

func (as *GasOracle) initDB(ctx context.Context, cfg config.Database) error {
	db, err := database.NewDB(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	as.db = db
	log.Info("Init database success")
	return nil
}

func (as *GasOracle) initSynchronizer(config *config.Config) error {
	for i := range config.RPCs {
		log.Info("Init synchronizer success", "chainId", config.RPCs[i].ChainId)
		rpcItem := config.RPCs[i]

		synchronizerTemp, err := synchronizer.NewOracleSynchronizer(as.db, as.ethClient[config.RPCs[i].ChainId], as.backOffset, rpcItem.ChainId, as.loopInternal, as.shutdown)
		if err != nil {
			log.Error("new oracle synchronizer fail", "err", err)
			return err
		}
		if as.Synchronizer == nil {
			as.Synchronizer = make(map[uint64]*synchronizer.OracleSynchronizer)
		}
		as.Synchronizer[rpcItem.ChainId] = synchronizerTemp
	}
	return nil
}
