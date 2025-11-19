package contracts

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"cs-projects-ink-eth-monitor/internal/client"
)

type AAveProtocolDataProvider struct {
	BaseContract
}

func NewAAveProtocolDataProvider(address common.Address) *AAveProtocolDataProvider {
	return &AAveProtocolDataProvider{
		BaseContract: NewBaseContract("aave_protocol_data_provider", address, TypeGetPaused),
	}
}

func (a *AAveProtocolDataProvider) Monitor(ctx context.Context, caller *client.ContractCaller) (float64, error) {
	methodID := crypto.Keccak256([]byte("getPaused(address)"))[:4]

	paramBytes, _ := hex.DecodeString(strings.TrimPrefix(L2WETH, "0x"))

	// 左填充到32字节
	paddedParam := make([]byte, 32)
	copy(paddedParam[32-len(paramBytes):], paramBytes)

	// 组合 methodID + 参数
	data := append(methodID, paddedParam...)

	// 调用合约
	paused, err := caller.CallBool(ctx, a.address.Hex(), data)
	if err != nil {
		return 0, err
	}

	// 返回指标值: true=1.0, false=0.0
	if paused {
		return 1.0, nil
	}
	return 0.0, nil
}
