package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/zap"

	"cs-projects-ink-eth-monitor/internal/config"
)

// Metrics 指标管理器
type Metrics struct {
	pusher         *push.Pusher
	logger         *zap.Logger
	gatewayURL     string
	jobName        string
	contractGauges map[string]prometheus.Gauge
	mu             sync.RWMutex
}

// NewMetrics 创建指标管理器
func NewMetrics(cfg *config.PrometheusConfig, logger *zap.Logger) *Metrics {
	m := &Metrics{
		logger:         logger,
		gatewayURL:     cfg.GatewayURL,
		jobName:        cfg.JobName,
		contractGauges: make(map[string]prometheus.Gauge),
	}

	// 创建pusher
	m.pusher = push.New(cfg.GatewayURL, cfg.JobName)

	return m
}

// getMetricName 根据链和合约名称返回自定义的指标名称
func getMetricName(chain, contractName string) string {
	key := fmt.Sprintf("%s_%s", chain, contractName)

	nameMap := map[string]string{
		"ethereum_super_chain_config":     "ink_eth_monitor_superchain_paused",
		"ethereum_ink_optimism_portal":    "ink_eth_monitor_optimism_portal_paused",
		"ethereum_l1_standard_bridge":     "ink_eth_monitor_standard_bridge_paused",
		"ink_aave_protocol_data_provider": "ink_eth_monitor_tydro_pool_paused",
		"ink_chaos_push_oracle":           "ink_eth_monitor_oracle_price_spread",
		"ink_variable_debt_InkWlWETH":     "ink_eth_monitor_remaining_supply",
	}

	if name, exists := nameMap[key]; exists {
		return name
	}

	// 默认使用标准命名方式
	return fmt.Sprintf("ink_eth_monitor_%s_%s", chain, contractName)
}

// RegisterContractMetric 注册合约指标
func (m *Metrics) RegisterContractMetric(chain, contractName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s_%s", chain, contractName)

	// 如果已经注册过，直接返回
	if _, exists := m.contractGauges[key]; exists {
		return
	}

	metricName := getMetricName(chain, contractName)

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: metricName,
		Help: fmt.Sprintf("Monitor metric for %s contract %s", chain, contractName),
		ConstLabels: prometheus.Labels{
			"chain":    chain,
			"contract": contractName,
		},
	})

	m.contractGauges[key] = gauge
	m.pusher.Collector(gauge)

	m.logger.Info("注册合约指标",
		zap.String("chain", chain),
		zap.String("contract", contractName),
		zap.String("metric_name", metricName),
	)
}

// SetContractMetric 设置合约指标值
func (m *Metrics) SetContractMetric(chain, contractName string, value float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s_%s", chain, contractName)

	if gauge, exists := m.contractGauges[key]; exists {
		gauge.Set(value)
		m.logger.Debug("设置指标值",
			zap.String("key", key),
			zap.Float64("value", value),
		)
	} else {
		m.logger.Warn("指标未注册",
			zap.String("key", key),
		)
	}
}

// Push 推送指标到Gateway
func (m *Metrics) Push() error {
	if err := m.pusher.Push(); err != nil {
		m.logger.Error("推送指标失败", zap.Error(err))
		return fmt.Errorf("推送指标到Prometheus Gateway失败: %w", err)
	}

	m.logger.Debug("成功推送指标到Prometheus Gateway")
	return nil
}

// Close 清理资源
func (m *Metrics) Close() error {
	// Prometheus pusher不需要特殊的关闭操作
	m.logger.Info("关闭指标管理器")
	return nil
}

// MetricValue 指标值
type MetricValue struct {
	Chain        string
	ContractName string
	Value        float64
}

// BatchSetMetrics 批量设置指标
func (m *Metrics) BatchSetMetrics(values []MetricValue) {
	for _, v := range values {
		m.SetContractMetric(v.Chain, v.ContractName, v.Value)
	}
}
