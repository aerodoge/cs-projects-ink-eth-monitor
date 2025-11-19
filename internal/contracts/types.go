package contracts

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"cs-projects-ink-eth-monitor/internal/client"
)

// 合约类型常量
const (
	TypePauseSimple     = "pause_simple"
	TypePauseIdentifier = "pause_with_identifier"
	TypePriceFeed       = "price_feed"
	TypeGetPaused       = "get_paused"
	TypeReserveCap      = "reserve_cap"
)

// 默认合约地址常量
// 注意: 这些地址可以通过配置文件覆盖
const (
	// Ethereum L1 合约地址
	DefaultL1SuperChainConfig  = "0x95703e0982140D16f8ebA6d158FccEde42f04a4C"
	DefaultL1StandardBridge    = "0x88FF1e5b602916615391F55854588EFcBB7663f0"
	DefaultL1InkOptimismPortal = "0x5d66C1782664115999C47c9fA5cd031f495D3e4F"
	L1WETH                     = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	// INK L2 合约地址
	DefaultL2AaveProtocolDataProvider = "0x96086C25d13943C80Ff9a19791a40Df6aFC08328"
	DefaultL2ChaosPushOracle          = "0x163131609562E578754aF12E998635BfCa56712C"
	DefaultL2VariableDebtInkWlWETH    = "0xc1457AcfBaD2332b07B7651A4Da3176E8F3Bc9E4"
	L2WETH                            = "0x4200000000000000000000000000000000000006"
)

// Account 合约接口
type Account interface {
	Name() string
	Address() common.Address
	Type() string
	// Monitor 执行监控并返回指标值
	Monitor(ctx context.Context, caller *client.ContractCaller) (float64, error)
}

// BaseContract 基础合约结构
type BaseContract struct {
	name     string
	address  common.Address
	typeName string
}

// NewBaseContract 创建基础合约
func NewBaseContract(name string, address common.Address, typeName string) BaseContract {
	return BaseContract{
		name:     name,
		address:  address,
		typeName: typeName,
	}
}

// Name 返回合约名称
func (b *BaseContract) Name() string {
	return b.name
}

// Address 返回合约地址
func (b *BaseContract) Address() common.Address {
	return b.address
}

// Type 返回合约类型
func (b *BaseContract) Type() string {
	return b.typeName
}
