package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"cs-projects-ink-eth-monitor/internal/config"
)

// ChainClient 链客户端接口
type ChainClient interface {
	// CallContract 调用合约方法
	CallContract(ctx context.Context, contract common.Address, method string, params []interface{}, result interface{}) error
	// Close 关闭客户端
	Close()
}

// EthClient Ethereum客户端
type EthClient struct {
	client *ethclient.Client
	logger *zap.Logger
	rpcURL string
}

// NewEthClient 创建Ethereum客户端
func NewEthClient(rpcURL string, logger *zap.Logger) (*EthClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接Ethereum RPC节点失败: %w", err)
	}

	logger.Info("成功连接到Ethereum RPC节点", zap.String("rpc_url", rpcURL))

	return &EthClient{
		client: client,
		logger: logger,
		rpcURL: rpcURL,
	}, nil
}

// CallContract 调用合约方法
func (c *EthClient) CallContract(ctx context.Context, contract common.Address, method string, params []interface{}, result interface{}) error {
	// 这里需要根据具体的ABI来构造调用
	// 为简化，我们使用直接的eth_call
	// 在实际使用中，需要加载具体的ABI

	// 构造方法签名的哈希
	// 由于没有ABI文件，这里简化处理
	// 实际项目中应该使用abigen生成的代码或者加载ABI

	c.logger.Debug("调用合约方法",
		zap.String("contract", contract.Hex()),
		zap.String("method", method),
	)

	// 这里返回一个占位符错误
	// 实际实现需要使用ABI
	return fmt.Errorf("需要实现具体的ABI调用逻辑")
}

// Close 关闭客户端
func (c *EthClient) Close() {
	c.client.Close()
	c.logger.Info("关闭Ethereum客户端")
}

// ContractCaller 合约调用器 - 简化版实现
type ContractCaller struct {
	client *ethclient.Client
	logger *zap.Logger
}

// NewContractCaller 创建合约调用器
func NewContractCaller(rpcURL string, logger *zap.Logger) (*ContractCaller, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接RPC节点失败: %w", err)
	}

	return &ContractCaller{
		client: client,
		logger: logger,
	}, nil
}

// CallBool 调用返回bool的方法
func (c *ContractCaller) CallBool(ctx context.Context, contractAddr string, data []byte) (bool, error) {
	msg := ethereum.CallMsg{
		To:   &common.Address{},
		Data: data,
	}
	copy(msg.To[:], common.HexToAddress(contractAddr).Bytes())

	result, err := c.client.CallContract(ctx, msg, nil)
	if err != nil {
		return false, fmt.Errorf("调用合约失败: %w", err)
	}

	// 解析bool结果 (32字节，最后一个字节为0或1)
	if len(result) < 32 {
		return false, fmt.Errorf("返回数据长度不足: %d", len(result))
	}

	return result[31] == 1, nil
}

// CallUint256 调用返回uint256的方法
func (c *ContractCaller) CallUint256(ctx context.Context, contractAddr string, data []byte) (*big.Int, error) {
	msg := ethereum.CallMsg{
		To:   &common.Address{},
		Data: data,
	}
	copy(msg.To[:], common.HexToAddress(contractAddr).Bytes())

	result, err := c.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("调用合约失败: %w", err)
	}

	// 解析uint256结果
	if len(result) < 32 {
		return nil, fmt.Errorf("返回数据长度不足: %d", len(result))
	}

	value := new(big.Int).SetBytes(result)
	return value, nil
}

// CallInt256 调用返回int256的方法（用于价格）
func (c *ContractCaller) CallInt256(ctx context.Context, contractAddr string, data []byte) (*big.Int, error) {
	// int256和uint256的解析方式相同，只是解释不同
	return c.CallUint256(ctx, contractAddr, data)
}

// CallRaw 调用合约方法并返回原始字节数据
func (c *ContractCaller) CallRaw(ctx context.Context, contractAddr string, data []byte) ([]byte, error) {
	msg := ethereum.CallMsg{
		To:   &common.Address{},
		Data: data,
	}
	copy(msg.To[:], common.HexToAddress(contractAddr).Bytes())

	result, err := c.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("调用合约失败: %w", err)
	}

	return result, nil
}

// Close 关闭客户端
func (c *ContractCaller) Close() {
	if c.client != nil {
		c.client.Close()
		c.logger.Info("关闭RPC客户端")
	}
}

// ClientManager 客户端管理器
type ClientManager struct {
	ethereumClient *ContractCaller
	inkClient      *ContractCaller
	logger         *zap.Logger
}

// NewClientManager 创建客户端管理器
func NewClientManager(cfg *config.Config, logger *zap.Logger) (*ClientManager, error) {
	// 创建Ethereum客户端
	ethClient, err := NewContractCaller(cfg.EthRPC, logger)
	if err != nil {
		return nil, fmt.Errorf("创建Ethereum客户端失败: %w", err)
	}
	logger.Info("成功创建Ethereum客户端", zap.String("rpc_url", cfg.EthRPC))

	// 创建INK客户端
	inkClient, err := NewContractCaller(cfg.InkRPC, logger)
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("创建INK客户端失败: %w", err)
	}
	logger.Info("成功创建INK客户端", zap.String("rpc_url", cfg.InkRPC))

	return &ClientManager{
		ethereumClient: ethClient,
		inkClient:      inkClient,
		logger:         logger,
	}, nil
}

// GetEthereumClient 获取Ethereum客户端
func (m *ClientManager) GetEthereumClient() *ContractCaller {
	return m.ethereumClient
}

// GetInkClient 获取INK客户端
func (m *ClientManager) GetInkClient() *ContractCaller {
	return m.inkClient
}

// Close 关闭所有客户端
func (m *ClientManager) Close() {
	if m.ethereumClient != nil {
		m.ethereumClient.Close()
	}
	if m.inkClient != nil {
		m.inkClient.Close()
	}
}
