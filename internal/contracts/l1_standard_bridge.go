package contracts

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"cs-projects-ink-eth-monitor/internal/client"
)

type InkStandardBridge struct {
	BaseContract
}

func NewInkStandardBridge(address common.Address) *InkStandardBridge {
	return &InkStandardBridge{
		BaseContract: NewBaseContract("l1_standard_bridge", address, TypePauseSimple),
	}
}

func (p *InkStandardBridge) Monitor(ctx context.Context, caller *client.ContractCaller) (float64, error) {
	methodID := crypto.Keccak256([]byte("paused()"))[:4]
	paused, err := caller.CallBool(ctx, p.address.Hex(), methodID)
	if err != nil {
		return 0, err
	}

	// 返回指标值: true=1.0, false=0.0
	if paused {
		return 1.0, nil
	}

	return 0.0, nil
}
