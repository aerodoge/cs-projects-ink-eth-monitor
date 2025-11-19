package contracts

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"cs-projects-ink-eth-monitor/internal/client"
)

type ChaosPushOracle struct {
	BaseContract
}

func NewChaosPushOracle(address common.Address) *ChaosPushOracle {
	return &ChaosPushOracle{
		BaseContract: NewBaseContract("chaos_push_oracle", address, TypePriceFeed),
	}
}

func (p *ChaosPushOracle) Monitor(ctx context.Context, caller *client.ContractCaller) (float64, error) {
	methodID := crypto.Keccak256([]byte("latestAnswer()"))[:4]
	price, err := caller.CallInt256(ctx, p.address.Hex(), methodID)
	if err != nil {
		return 0, err
	}

	// 将价格转换为float64 (除以10^8，因为Chainlink价格有8位小数)
	priceFloat := new(big.Float).SetInt(price)
	divisor := new(big.Float).SetFloat64(100000000) // 10^8
	priceFloat.Quo(priceFloat, divisor)
	value, _ := priceFloat.Float64()

	return value, nil
}
