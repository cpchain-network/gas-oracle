package synchronizer

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/cpchain-network/gas-oracle/common/tasks"
	"github.com/cpchain-network/gas-oracle/database"
	"github.com/cpchain-network/gas-oracle/synchronizer/node"
)

type OracleSynchronizer struct {
	loopInternal   time.Duration
	db             *database.DB
	ethClient      node.EthClient
	blockOffset    uint64
	chainId        uint64
	nativeToken    string
	decimal        uint8
	stopped        atomic.Bool
	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	tasks          tasks.Group
}

func (os *OracleSynchronizer) Stop(ctx context.Context) error {
	os.stopped.Store(true)
	return nil
}

func (os *OracleSynchronizer) Stopped() bool {
	return os.stopped.Load()
}

func NewOracleSynchronizer(db *database.DB, client node.EthClient, blockOffset uint64, chainId uint64, nativeToken string, decimal uint8, loopInternal time.Duration, shutdown context.CancelCauseFunc) (*OracleSynchronizer, error) {

	resCtx, resCancel := context.WithCancel(context.Background())

	return &OracleSynchronizer{
		loopInternal: loopInternal,
		db:           db,
		chainId:      chainId,
		nativeToken:  nativeToken,
		decimal:      decimal,
		ethClient:    client,
		blockOffset:  blockOffset,
		tasks: tasks.Group{HandleCrit: func(err error) {
			shutdown(fmt.Errorf("critical error in selaginella processor: %w", err))
		}},
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
	}, nil
}

func (os *OracleSynchronizer) Start(ctx context.Context) error {
	l1FeeTicker := time.NewTicker(os.loopInternal)
	os.tasks.Go(func() error {
		for range l1FeeTicker.C {
			fee, err := os.processTokenPrice(os.chainId)
			if err != nil {
				log.Error("process token price error", "err", err)
				log.Error(err.Error())
			}
			log.Info("get gas fee", "fee", fee, "chainId", os.chainId)
			gasFee := &database.GasFee{
				GUID:       uuid.New(),
				ChainId:    big.NewInt(int64(os.chainId)),
				Decimal:    os.decimal,
				TokenName:  os.nativeToken,
				PredictFee: fee.String(),
				Timestamp:  uint64(time.Now().Unix()),
			}
			err = os.db.GasFee.StoreOrUpdateGasFee(gasFee)
			if err != nil {
				log.Error("Oracle synchronizer store or update gas fee fail", "err", err)
				return err
			}
		}
		return nil
	})
	return nil
}

func (os *OracleSynchronizer) processTokenPrice(chainId uint64) (*big.Int, error) {
	var gasPrice *big.Int
	var transactionFee *big.Int
	var blockFee = big.NewInt(0)
	var fee = big.NewInt(0)

	latestBlockN, err := os.ethClient.GetLatestBlock(context.Background())
	if err != nil {
		log.Error("failed to get l1 latest block number", "err", err)
		return nil, err
	}
	log.Info("start handle block fee", "blockOffset", os.blockOffset, "latestBlockN", latestBlockN.String())
	for i := 0; i < int(os.blockOffset); i++ {
		blockNumber := int(latestBlockN.Int64()) - i
		txs, baseFee, err := os.ethClient.BlockDetailByNumber(context.Background(), big.NewInt(int64(blockNumber)))
		if err != nil {
			log.Error("failed to get block", "blockNum", blockNumber, "err", err)
			return nil, err
		}
		log.Info("successfully get block info", "block_num", blockNumber, "tx_len", len(txs))

		for _, tx := range txs {
			receipt, err := os.ethClient.TxReceiptDetailByHash(context.Background(), common.HexToHash(tx))
			if err != nil {
				log.Error("failed to get transaction receipt", "tx_hash", tx, "err", err)
				return nil, err
			}
			if receipt.Type == types.DynamicFeeTxType {
				gasPrice = receipt.EffectiveGasPrice
				transactionFee = new(big.Int).Add(gasPrice, baseFee)
				transactionFee.Mul(transactionFee, new(big.Int).SetUint64(receipt.GasUsed))
			} else {
				gasPrice = receipt.EffectiveGasPrice
				transactionFee = new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(receipt.GasUsed))
			}
			blockFee = new(big.Int).Add(blockFee, transactionFee)
		}

		blockFee = new(big.Int).Div(blockFee, big.NewInt(int64(len(txs))))
		fee = new(big.Int).Add(fee, new(big.Int).Div(blockFee, big.NewInt(int64(os.blockOffset))))
	}
	log.Info("successfully get estimated fee", "chainId", chainId, "fee", fee)
	return fee, nil
}
