package contracts

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"cs-projects-ink-eth-monitor/internal/client"
)

const (
	EthRPC = "https://eth-mainnet.g.alchemy.com/v2/jP0h5UEZoR7Wpww9tnNKPGihmNwEkECH"
	InkRPC = "https://ink-mainnet.g.alchemy.com/v2/jP0h5UEZoR7Wpww9tnNKPGihmNwEkECH"
)

// getTestLogger 获取测试用的 logger
func getTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// getTestEthClient 获取测试用的 Ethereum 客户端
func getTestEthClient(t *testing.T) *client.ContractCaller {
	logger := getTestLogger()
	caller, err := client.NewContractCaller(EthRPC, logger)
	if err != nil {
		t.Fatalf("创建 Ethereum 客户端失败: %v", err)
	}
	return caller
}

// getTestInkClient 获取测试用的 INK 客户端
func getTestInkClient(t *testing.T) *client.ContractCaller {
	logger := getTestLogger()
	caller, err := client.NewContractCaller(InkRPC, logger)
	if err != nil {
		t.Fatalf("创建 INK 客户端失败: %v", err)
	}
	return caller
}

// TestL1SuperChainConfig 测试 L1 SuperChainConfig 真实监控
func TestL1SuperChainConfig(t *testing.T) {
	caller := getTestEthClient(t)
	address := common.HexToAddress(DefaultL1SuperChainConfig)
	contract := NewSuperChainConfig(address)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	value, err := contract.Monitor(ctx, caller)
	if err != nil {
		t.Fatalf("Monitor() 失败: %v", err)
	}

	// 验证返回值应该是 0.0 或 1.0
	if value != 0.0 && value != 1.0 {
		t.Errorf("期望返回值为 0.0 或 1.0, 得到 %v", value)
	}

	t.Logf("SuperChainConfig 暂停状态: %v", value)
}

// TestL1InkOptimismPortal 测试 L1 InkOptimismPortal 真实监控
func TestL1InkOptimismPortal(t *testing.T) {
	caller := getTestEthClient(t)
	address := common.HexToAddress(DefaultL1InkOptimismPortal)
	contract := NewInkOptimismPortal(address)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	value, err := contract.Monitor(ctx, caller)
	if err != nil {
		t.Fatalf("Monitor() 失败: %v", err)
	}

	// 验证返回值应该是 0.0 或 1.0
	if value != 0.0 && value != 1.0 {
		t.Errorf("期望返回值为 0.0 或 1.0, 得到 %v", value)
	}

	t.Logf("InkOptimismPortal 暂停状态: %v", value)
}

// TestL1StandardBridge 测试 L1 StandardBridge 真实监控
func TestL1StandardBridge(t *testing.T) {
	caller := getTestEthClient(t)
	address := common.HexToAddress(DefaultL1StandardBridge)
	contract := NewInkStandardBridge(address)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	value, err := contract.Monitor(ctx, caller)
	if err != nil {
		t.Fatalf("Monitor() 失败: %v", err)
	}

	// 验证返回值应该是 0.0 或 1.0
	if value != 0.0 && value != 1.0 {
		t.Errorf("期望返回值为 0.0 或 1.0, 得到 %v", value)
	}

	t.Logf("StandardBridge 暂停状态: %v", value)
}

// TestL2AAveProtocolDataProvider 测试 L2 AAve 真实监控
func TestL2AAveProtocolDataProvider(t *testing.T) {
	caller := getTestInkClient(t)
	address := common.HexToAddress(DefaultL2AaveProtocolDataProvider)
	contract := NewAAveProtocolDataProvider(address)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	value, err := contract.Monitor(ctx, caller)
	if err != nil {
		t.Fatalf("Monitor() 失败: %v", err)
	}

	// 验证返回值应该是 0.0 或 1.0
	if value != 0.0 && value != 1.0 {
		t.Errorf("期望返回值为 0.0 或 1.0, 得到 %v", value)
	}

	t.Logf("AAveProtocolDataProvider 暂停状态: %v", value)
}

// TestL2ChaosPushOracle 测试 L2 ChaosPushOracle 真实监控
func TestL2ChaosPushOracle(t *testing.T) {
	caller := getTestInkClient(t)
	address := common.HexToAddress(DefaultL2ChaosPushOracle)
	contract := NewChaosPushOracle(address)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	inkPrice, err := contract.Monitor(ctx, caller)
	if err != nil {
		t.Fatalf("Monitor() 失败: %v", err)
	}
	fmt.Printf("ink price: %v\n", inkPrice)

	// 创建以太坊主网 Chainlink ETH/USD 预言机实例
	chainlinkAddr := common.HexToAddress("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	chainlink := NewChaosPushOracle(chainlinkAddr)

	// 获取以太坊主网的价格
	ethPrice, err := chainlink.Monitor(ctx, getTestEthClient(t))
	if err != nil {
		fmt.Errorf("获取ETH主网价格失败: %w", err)
	}
	fmt.Printf("eth price: %v\n", ethPrice)
	// 计算价格偏差
	var deviation float64
	if ethPrice != 0 {
		deviation = (inkPrice - ethPrice) / ethPrice
		if deviation < 0 {
			deviation = -deviation // 取绝对值
		}
	}

	t.Logf("价差: %v", deviation)
}

// TestL2InkWLWEth 测试 L2 InkWLWEth 真实监控
func TestL2InkWLWEth(t *testing.T) {
	caller := getTestInkClient(t)
	address := common.HexToAddress(DefaultL2VariableDebtInkWlWETH)
	contract := NewInkWLWEth(address)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	value, err := contract.Monitor(ctx, caller)
	if err != nil {
		t.Fatalf("Monitor() 失败: %v", err)
	}

	// 验证返回值应该大于等于 0
	if value < 0 {
		t.Errorf("期望储备上限 >= 0, 得到 %v", value)
	}

	t.Logf("InkWLWEth 储备上限: %v", value)
}

// TestAllContracts_Properties 测试所有合约的属性方法
func TestAllContracts_Properties(t *testing.T) {
	tests := []struct {
		name     string
		contract Account
		wantName string
		wantType string
	}{
		{
			name:     "SuperChainConfig",
			contract: NewSuperChainConfig(common.HexToAddress(DefaultL1SuperChainConfig)),
			wantName: "super_chain_config",
			wantType: TypePauseSimple,
		},
		{
			name:     "InkOptimismPortal",
			contract: NewInkOptimismPortal(common.HexToAddress(DefaultL1InkOptimismPortal)),
			wantName: "ink_optimism_portal",
			wantType: TypePauseSimple,
		},
		{
			name:     "InkStandardBridge",
			contract: NewInkStandardBridge(common.HexToAddress(DefaultL1StandardBridge)),
			wantName: "l1_standard_bridge",
			wantType: TypePauseSimple,
		},
		{
			name:     "AAveProtocolDataProvider",
			contract: NewAAveProtocolDataProvider(common.HexToAddress(DefaultL2AaveProtocolDataProvider)),
			wantName: "aave_protocol_data_provider",
			wantType: TypeGetPaused,
		},
		{
			name:     "ChaosPushOracle",
			contract: NewChaosPushOracle(common.HexToAddress(DefaultL2ChaosPushOracle)),
			wantName: "chaos_push_oracle",
			wantType: TypePriceFeed,
		},
		{
			name:     "InkWLWEth",
			contract: NewInkWLWEth(common.HexToAddress(DefaultL2VariableDebtInkWlWETH)),
			wantName: "variable_debt_InkWlWETH",
			wantType: TypeReserveCap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.contract.Name() != tt.wantName {
				t.Errorf("Name() = %v, 期望 %v", tt.contract.Name(), tt.wantName)
			}
			if tt.contract.Type() != tt.wantType {
				t.Errorf("Type() = %v, 期望 %v", tt.contract.Type(), tt.wantType)
			}
			if tt.contract.Address() == (common.Address{}) {
				t.Error("Address() 返回零地址")
			}
		})
	}
}
