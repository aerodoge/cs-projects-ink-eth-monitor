package emergency

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"

	"cs-projects-ink-eth-monitor/internal/config"
	"cs-projects-ink-eth-monitor/internal/contracts"
)

// Manager 应急响应管理器
type Manager struct {
	cfg             *config.EmergencyConfig
	logger          *zap.Logger
	delegate        *contracts.Delegate
	triggered       bool
	lastTriggerTime time.Time
	mu              sync.Mutex
}

// NewManager 创建应急响应管理器
func NewManager(cfg *config.EmergencyConfig, inkRPC string, logger *zap.Logger) (*Manager, error) {
	if !cfg.Enabled {
		logger.Info("应急响应功能未启用")
		return &Manager{
			cfg:    cfg,
			logger: logger,
		}, nil
	}

	// 验证配置
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("应急响应配置错误: private_key 不能为空")
	}
	if cfg.SafeAddress == "" {
		return nil, fmt.Errorf("应急响应配置错误: safe_address 不能为空")
	}
	if cfg.ArgusAddress == "" {
		return nil, fmt.Errorf("应急响应配置错误: argus_address 不能为空")
	}
	if cfg.WithdrawAmount == "" {
		return nil, fmt.Errorf("应急响应配置错误: withdraw_amount 不能为空")
	}

	// 创建 Delegate
	delegate := contracts.NewDelegate(
		inkRPC,
		cfg.PrivateKey,
		cfg.SafeAddress,
		cfg.ArgusAddress,
	)

	logger.Info("应急响应管理器已启用",
		zap.String("safe_address", cfg.SafeAddress),
		zap.String("argus_address", cfg.ArgusAddress),
		zap.String("withdraw_amount", cfg.WithdrawAmount),
	)

	return &Manager{
		cfg:      cfg,
		logger:   logger,
		delegate: delegate,
	}, nil
}

// CheckAlert 检查是否触发告警并执行应急响应
func (m *Manager) CheckAlert(metricName string, value float64) error {
	if !m.cfg.Enabled {
		return nil
	}

	// 判断是否触发告警
	shouldTrigger := false
	alertReason := ""

	switch metricName {
	case "ink_eth_monitor_superchain_paused":
		if value == 1.0 {
			shouldTrigger = true
			alertReason = "SuperChain 合约已暂停"
		}
	case "ink_eth_monitor_optimism_portal_paused":
		if value == 1.0 {
			shouldTrigger = true
			alertReason = "Optimism Portal 合约已暂停"
		}
	case "ink_eth_monitor_standard_bridge_paused":
		if value == 1.0 {
			shouldTrigger = true
			alertReason = "Standard Bridge 合约已暂停"
		}
	case "ink_eth_monitor_tydro_pool_paused":
		if value == 1.0 {
			shouldTrigger = true
			alertReason = "Tydro Pool 合约已暂停"
		}
	case "ink_eth_monitor_oracle_price_spread":
		if value > 0.05 {
			shouldTrigger = true
			alertReason = fmt.Sprintf("价格偏差过大: %.2f%% (超过5%%)", value*100)
		}
	case "ink_eth_monitor_remaining_supply":
		if value < 2500 {
			shouldTrigger = true
			alertReason = fmt.Sprintf("剩余容量不足: %.2f tokens (低于2500)", value)
		}
	}

	if shouldTrigger {
		return m.executeEmergencyWithdraw(alertReason)
	}

	return nil
}

// executeEmergencyWithdraw 执行应急提款
func (m *Manager) executeEmergencyWithdraw(reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已经触发过（防止重复执行）
	if m.triggered {
		m.logger.Warn("应急响应已触发过，跳过本次执行",
			zap.String("reason", reason),
			zap.Time("last_trigger_time", m.lastTriggerTime),
		)
		return nil
	}

	m.logger.Warn("🚨 触发应急响应！开始执行提款操作...",
		zap.String("reason", reason),
		zap.String("withdraw_amount", m.cfg.WithdrawAmount),
	)

	// 解析提款金额
	amount, ok := new(big.Int).SetString(m.cfg.WithdrawAmount, 10)
	if !ok {
		return fmt.Errorf("无法解析提款金额: %s", m.cfg.WithdrawAmount)
	}

	// 执行提款
	err := m.delegate.WithdrawETHFromGatewayV3(amount)
	if err != nil {
		m.logger.Error("应急提款失败", zap.Error(err))
		return fmt.Errorf("应急提款失败: %w", err)
	}

	// 标记已触发
	m.triggered = true
	m.lastTriggerTime = time.Now()

	m.logger.Info("✅ 应急提款执行成功",
		zap.String("reason", reason),
		zap.String("amount", m.cfg.WithdrawAmount),
		zap.Time("trigger_time", m.lastTriggerTime),
	)

	return nil
}

// IsTriggered 检查是否已触发
func (m *Manager) IsTriggered() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.triggered
}

// Reset 重置触发状态（用于测试或手动恢复）
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.triggered = false
	m.lastTriggerTime = time.Time{}
	m.logger.Info("应急响应状态已重置")
}

// Close 关闭应急响应管理器
func (m *Manager) Close() error {
	m.logger.Info("关闭应急响应管理器")
	return nil
}
