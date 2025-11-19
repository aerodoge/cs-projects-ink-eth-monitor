package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"cs-projects-ink-eth-monitor/internal/client"
	"cs-projects-ink-eth-monitor/internal/config"
	"cs-projects-ink-eth-monitor/internal/contracts"
	"cs-projects-ink-eth-monitor/internal/metrics"
	"cs-projects-ink-eth-monitor/pkg/retry"
)

// Monitor 监控器
type Monitor struct {
	cfg           *config.Config
	clientManager *client.ClientManager
	metrics       *metrics.Metrics
	logger        *zap.Logger
	stopChan      chan struct{}
	ethAccounts   []contracts.Account
	inkAccounts   []contracts.Account
}

// NewMonitor 创建监控器
func NewMonitor(
	cfg *config.Config,
	clientManager *client.ClientManager,
	metricsManager *metrics.Metrics,
	logger *zap.Logger,
) *Monitor {
	// 使用配置中的地址，如果未配置则使用默认值
	l1SuperChainConfig := getAddressOrDefault(cfg.Contracts.L1.SuperChainConfig, contracts.DefaultL1SuperChainConfig)
	l1StandardBridge := getAddressOrDefault(cfg.Contracts.L1.StandardBridge, contracts.DefaultL1StandardBridge)
	l1InkOptimismPortal := getAddressOrDefault(cfg.Contracts.L1.InkOptimismPortal, contracts.DefaultL1InkOptimismPortal)
	l2AaveProtocolDataProvider := getAddressOrDefault(cfg.Contracts.L2.AaveProtocolDataProvider, contracts.DefaultL2AaveProtocolDataProvider)
	l2ChaosPushOracle := getAddressOrDefault(cfg.Contracts.L2.ChaosPushOracle, contracts.DefaultL2ChaosPushOracle)
	l2VariableDebtInkWlWETH := getAddressOrDefault(cfg.Contracts.L2.VariableDebtInkWlWETH, contracts.DefaultL2VariableDebtInkWlWETH)

	return &Monitor{
		cfg:           cfg,
		clientManager: clientManager,
		metrics:       metricsManager,
		logger:        logger,
		stopChan:      make(chan struct{}),
		ethAccounts: []contracts.Account{
			contracts.NewSuperChainConfig(common.HexToAddress(l1SuperChainConfig)),
			contracts.NewInkOptimismPortal(common.HexToAddress(l1InkOptimismPortal)),
			contracts.NewInkStandardBridge(common.HexToAddress(l1StandardBridge)),
		},
		inkAccounts: []contracts.Account{
			contracts.NewAAveProtocolDataProvider(common.HexToAddress(l2AaveProtocolDataProvider)),
			contracts.NewChaosPushOracle(common.HexToAddress(l2ChaosPushOracle)),
			contracts.NewInkWLWEth(common.HexToAddress(l2VariableDebtInkWlWETH)),
		},
	}
}

// getAddressOrDefault 返回配置的地址，如果为空则返回默认值
func getAddressOrDefault(configAddr, defaultAddr string) string {
	if configAddr != "" {
		return configAddr
	}
	return defaultAddr
}

// Start 启动监控
func (m *Monitor) Start(ctx context.Context) error {
	m.logger.Info("启动监控服务")

	// 注册所有指标
	m.registerMetrics()

	// 启动轮询
	ticker := time.NewTicker(m.cfg.Monitor.GetPollDuration())
	defer ticker.Stop()

	// 立即执行一次
	m.pollAll(ctx)

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("监控服务收到停止信号")
			return ctx.Err()
		case <-m.stopChan:
			m.logger.Info("监控服务停止")
			return nil
		case <-ticker.C:
			m.pollAll(ctx)
		}
	}
}

// Stop 停止监控
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// registerMetrics 注册所有指标
func (m *Monitor) registerMetrics() {
	// 注册Ethereum合约指标
	for _, contract := range m.ethAccounts {
		m.metrics.RegisterContractMetric("ethereum", contract.Name())
	}

	// 注册INK合约指标
	for _, contract := range m.inkAccounts {
		m.metrics.RegisterContractMetric("ink", contract.Name())
	}

	m.logger.Info("完成指标注册")
}

// pollAll 轮询所有合约
func (m *Monitor) pollAll(ctx context.Context) {
	m.logger.Debug("开始轮询所有合约")

	// 轮询Ethereum合约
	for _, contract := range m.ethAccounts {
		m.pollEthereumContract(ctx, contract)
	}

	// 轮询INK合约
	for _, contract := range m.inkAccounts {
		m.pollInkContract(ctx, contract)
	}

	// 推送指标到Prometheus Gateway
	if err := m.metrics.Push(); err != nil {
		m.logger.Error("推送指标失败", zap.Error(err))
	}
}

// pollEthereumContract 轮询Ethereum合约
func (m *Monitor) pollEthereumContract(ctx context.Context, contract contracts.Account) {
	err := retry.Do(ctx, func() error {
		return m.checkEthereumContract(ctx, contract)
	}, m.cfg.Monitor.RetryTimes, m.cfg.Monitor.GetRetryDelay(), m.logger)

	if err != nil {
		m.logger.Error("检查Ethereum合约失败",
			zap.String("contract", contract.Address().Hex()),
			zap.String("name", contract.Name()),
			zap.Error(err),
		)
	}
}

// pollInkContract 轮询INK合约
func (m *Monitor) pollInkContract(ctx context.Context, contract contracts.Account) {
	err := retry.Do(ctx, func() error {
		return m.checkInkContract(ctx, contract)
	}, m.cfg.Monitor.RetryTimes, m.cfg.Monitor.GetRetryDelay(), m.logger)

	if err != nil {
		m.logger.Error("检查INK合约失败",
			zap.String("contract", contract.Address().Hex()),
			zap.String("name", contract.Name()),
			zap.Error(err),
		)
	}
}

// checkEthereumContract 检查Ethereum合约
func (m *Monitor) checkEthereumContract(ctx context.Context, contract contracts.Account) error {
	// 调用合约的Monitor方法获取指标值
	value, err := contract.Monitor(ctx, m.clientManager.GetEthereumClient())
	if err != nil {
		return fmt.Errorf("监控合约失败: %w", err)
	}

	// 设置指标值
	m.metrics.SetContractMetric("ethereum", contract.Name(), contract.Type(), value)

	m.logger.Info("检查Ethereum合约",
		zap.String("contract", contract.Name()),
		zap.String("type", contract.Type()),
		zap.Float64("value", value),
	)

	return nil
}

// checkInkContract 检查INK合约
func (m *Monitor) checkInkContract(ctx context.Context, contract contracts.Account) error {
	// 特殊处理：ChaosPushOracle 需要跨链价格比较
	if contract.Type() == contracts.TypePriceFeed && contract.Name() == "chaos_push_oracle" {
		return m.checkPriceFeedDeviation(ctx, contract)
	}

	// 其他合约正常处理
	value, err := contract.Monitor(ctx, m.clientManager.GetInkClient())
	if err != nil {
		return fmt.Errorf("监控合约失败: %w", err)
	}

	// 设置指标值
	m.metrics.SetContractMetric("ink", contract.Name(), contract.Type(), value)

	m.logger.Info("检查INK合约",
		zap.String("contract", contract.Name()),
		zap.String("type", contract.Type()),
		zap.Float64("value", value),
	)

	return nil
}

// checkPriceFeedDeviation 检查价格源偏差（跨链比较）
func (m *Monitor) checkPriceFeedDeviation(ctx context.Context, contract contracts.Account) error {
	// 1. 获取 INK 链上的价格
	inkPrice, err := contract.Monitor(ctx, m.clientManager.GetInkClient())
	if err != nil {
		return fmt.Errorf("获取INK链价格失败: %w", err)
	}

	// 2. 创建以太坊主网 Chainlink ETH/USD 预言机实例
	chainlinkAddr := common.HexToAddress("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	chainlink := contracts.NewChaosPushOracle(chainlinkAddr)

	// 3. 获取以太坊主网的价格
	ethPrice, err := chainlink.Monitor(ctx, m.clientManager.GetEthereumClient())
	if err != nil {
		return fmt.Errorf("获取ETH主网价格失败: %w", err)
	}

	// 4. 计算价格偏差
	var deviation float64
	if ethPrice != 0 {
		deviation = (inkPrice - ethPrice) / ethPrice
		if deviation < 0 {
			deviation = -deviation // 取绝对值
		}
	}

	// 5. 判断偏差是否超过阈值（5%）
	value := 0.0
	if deviation > 0.05 {
		value = 1.0 // 告警：价格偏差过大
	}

	// 6. 设置指标值
	m.metrics.SetContractMetric("ink", contract.Name(), value)

	m.logger.Info("检查价格源偏差",
		zap.String("contract", contract.Name()),
		zap.Float64("ink_price", inkPrice),
		zap.Float64("eth_price", ethPrice),
		zap.Float64("deviation", deviation),
		zap.Float64("alert_value", value),
	)

	return nil
}
