package contracts

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"cs-projects-ink-eth-monitor/internal/client"
)

type InkWLWEth struct {
	BaseContract
}

func NewInkWLWEth(address common.Address) *InkWLWEth {
	return &InkWLWEth{
		BaseContract: NewBaseContract("variable_debt_InkWlWETH", address, TypeReserveCap),
	}
}

func (p *InkWLWEth) Monitor(ctx context.Context, caller *client.ContractCaller) (float64, error) {
	// 1. 调用 AAveProtocolDataProvider.getReserveCaps(address) 获取 supplyCap
	aaveAddr := DefaultL2AaveProtocolDataProvider
	assetAddr := "0x4200000000000000000000000000000000000006"

	// 构造 getReserveCaps(address) 调用
	methodID := crypto.Keccak256([]byte("getReserveCaps(address)"))[:4]
	assetBytes, _ := hex.DecodeString(strings.TrimPrefix(assetAddr, "0x"))
	paddedAsset := make([]byte, 32)
	copy(paddedAsset[32-len(assetBytes):], assetBytes)
	data := append(methodID, paddedAsset...)

	// 调用合约获取原始返回数据
	result, err := caller.CallRaw(ctx, aaveAddr, data)
	if err != nil {
		return 0, fmt.Errorf("调用 getReserveCaps 失败: %w", err)
	}

	// 解析返回值：(uint256 borrowCap, uint256 supplyCap)
	if len(result) < 64 {
		return 0, fmt.Errorf("返回数据长度不足: %d", len(result))
	}

	supplyCap := new(big.Int).SetBytes(result[32:64])

	// 2. 调用 InkWLWEth.totalSupply() 获取当前供应量
	totalSupplyMethodID := crypto.Keccak256([]byte("totalSupply()"))[:4]
	totalSupply, err := caller.CallUint256(ctx, p.address.Hex(), totalSupplyMethodID)
	if err != nil {
		return 0, fmt.Errorf("调用 totalSupply 失败: %w", err)
	}

	// 将 totalSupply 从 wei 转换为 token 单位（除以 10^18）
	totalSupplyFloat := new(big.Float).SetInt(totalSupply)
	divisor := new(big.Float).SetFloat64(1e18)
	totalSupplyFloat.Quo(totalSupplyFloat, divisor)
	totalSupplyInTokens, _ := totalSupplyFloat.Float64()

	// 3. 计算剩余容量：supplyCap - totalSupply (都是 token 单位)
	supplyCapFloat := new(big.Float).SetInt(supplyCap)
	remaining := supplyCapFloat.Sub(supplyCapFloat, new(big.Float).SetFloat64(totalSupplyInTokens))
	value, _ := remaining.Float64()

	return value, nil
}
