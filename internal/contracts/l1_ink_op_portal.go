package contracts

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"cs-projects-ink-eth-monitor/internal/client"
)

// InkOptimismPortal L1 Ink Optimism Portal 合约
type InkOptimismPortal struct {
	BaseContract
}

// NewInkOptimismPortal 创建 InkOptimismPortal 实例
func NewInkOptimismPortal(address common.Address) *InkOptimismPortal {
	return &InkOptimismPortal{
		BaseContract: NewBaseContract("ink_optimism_portal", address, TypePauseSimple),
	}
}

// Monitor 监控合约的暂停状态
// Keccak256:
//  1. 只包含参数类型，不包含返回值类型
//  2. 参数类型之间用逗号分隔，无空格
//  3. 不包含参数名，只有类型
//  4. 不包含 view、pure、external 等修饰符
func (p *InkOptimismPortal) Monitor(ctx context.Context, caller *client.ContractCaller) (float64, error) {
	// 构造 paused() 方法的 calldata
	methodID := crypto.Keccak256([]byte("paused()"))[:4]

	// 调用合约
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
