package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Log        LogConfig        `mapstructure:"log"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Monitor    MonitorConfig    `mapstructure:"monitor"`
	Contracts  ContractsConfig  `mapstructure:"contracts"`
	EthRPC     string           `mapstructure:"eth_rpc"`
	InkRPC     string           `mapstructure:"ink_rpc"`
}

// ContractsConfig 合约地址配置（可选）
type ContractsConfig struct {
	L1 L1ContractsConfig `mapstructure:"l1"`
	L2 L2ContractsConfig `mapstructure:"l2"`
}

// L1ContractsConfig L1合约地址配置
type L1ContractsConfig struct {
	SuperChainConfig  string `mapstructure:"superchain_config"`
	StandardBridge    string `mapstructure:"standard_bridge"`
	InkOptimismPortal string `mapstructure:"ink_optimism_portal"`
}

// L2ContractsConfig L2合约地址配置
type L2ContractsConfig struct {
	AaveProtocolDataProvider string `mapstructure:"aave_protocol_data_provider"`
	ChaosPushOracle          string `mapstructure:"chaos_push_oracle"`
	VariableDebtInkWlWETH    string `mapstructure:"variable_debt_inkwlweth"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	GatewayURL   string `mapstructure:"gateway_url"`
	JobName      string `mapstructure:"job_name"`
	PushInterval int    `mapstructure:"push_interval"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	PollInterval int `mapstructure:"poll_interval"`
	RetryTimes   int `mapstructure:"retry_times"`
	RetryDelay   int `mapstructure:"retry_delay"`
}

// ChainConfig 链配置
type ChainConfig struct {
	RpcURL    string           `mapstructure:"rpc_url"`
	Contracts []ContractConfig `mapstructure:"contracts"`
}

// ContractConfig 合约配置
type ContractConfig struct {
	Address      string        `mapstructure:"address"`
	Name         string        `mapstructure:"name"`
	Type         string        `mapstructure:"type"`
	Method       string        `mapstructure:"method"`
	MethodParams []interface{} `mapstructure:"method_params"`
	Alert        *AlertConfig  `mapstructure:"alert"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	Type           string      `mapstructure:"type"`
	Threshold      interface{} `mapstructure:"threshold"` // 可以是float64或string
	CompareWith    string      `mapstructure:"compare_with"`
	CompareAddress string      `mapstructure:"compare_address"`
	CompareMethod  string      `mapstructure:"compare_method"`
}

// GetThresholdFloat 获取浮点数阈值
func (a *AlertConfig) GetThresholdFloat() float64 {
	if v, ok := a.Threshold.(float64); ok {
		return v
	}
	if v, ok := a.Threshold.(int); ok {
		return float64(v)
	}
	return 0
}

// GetThresholdString 获取字符串阈值
func (a *AlertConfig) GetThresholdString() string {
	if v, ok := a.Threshold.(string); ok {
		return v
	}
	return ""
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.EthRPC == "" {
		return fmt.Errorf("eth_rpc 不能为空")
	}
	if c.InkRPC == "" {
		return fmt.Errorf("ink_rpc 不能为空")
	}
	if c.Prometheus.GatewayURL == "" {
		return fmt.Errorf("prometheus.gateway_url 不能为空")
	}
	if c.Monitor.PollInterval <= 0 {
		return fmt.Errorf("monitor.poll_interval 必须大于0")
	}
	return nil
}

// GetPollDuration 获取轮询间隔时间
func (c *MonitorConfig) GetPollDuration() time.Duration {
	return time.Duration(c.PollInterval) * time.Second
}

// GetRetryDelay 获取重试延迟时间
func (c *MonitorConfig) GetRetryDelay() time.Duration {
	return time.Duration(c.RetryDelay) * time.Second
}

// GetPushDuration 获取推送间隔时间
func (c *PrometheusConfig) GetPushDuration() time.Duration {
	return time.Duration(c.PushInterval) * time.Second
}
