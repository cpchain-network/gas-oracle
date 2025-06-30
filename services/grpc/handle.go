package grpc

import (
	"context"
	"strconv"

	"github.com/ethereum/go-ethereum/log"

	"github.com/cpchain-network/gas-oracle/proto/gasfee"
)

func (ms *GasFeeRpcService) GetGasFeeByChainId(ctx context.Context, in *gasfee.GasFeeRequest) (*gasfee.GasFeeResponse, error) {
	gasFee, err := ms.db.GasFee.QueryGasFees(strconv.FormatUint(in.ChainId, 10))
	if err != nil {
		log.Error("Query gas fee", "err", err)
		return nil, err
	}
	return &gasfee.GasFeeResponse{
		ReturnCode: 100,
		Message:    "get gas fee success",
		GasFee:     gasFee.GasFee.Uint64(),
		BlockFee:   0,
	}, nil
}
