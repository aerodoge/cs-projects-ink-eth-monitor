package contracts

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"cs-projects-ink-eth-monitor/internal/client"
)

type SuperChainConfig struct {
	BaseContract
}

func NewSuperChainConfig(address common.Address) *SuperChainConfig {
	return &SuperChainConfig{
		BaseContract: NewBaseContract("super_chain_config", address, TypePauseSimple),
	}
}

func (s *SuperChainConfig) Monitor(ctx context.Context, caller *client.ContractCaller) (float64, error) {
	// 构造 paused() 方法的 calldata
	methodID := crypto.Keccak256([]byte("paused()"))[:4]

	// 调用合约
	paused, err := caller.CallBool(ctx, s.address.Hex(), methodID)
	if err != nil {
		return 0, err
	}

	// 返回指标值: true=1.0, false=0.0
	if paused {
		return 1.0, nil
	}
	return 0.0, nil
}
