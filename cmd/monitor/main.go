package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"cs-projects-ink-eth-monitor/internal/client"
	"cs-projects-ink-eth-monitor/internal/config"
	"cs-projects-ink-eth-monitor/internal/emergency"
	"cs-projects-ink-eth-monitor/internal/logger"
	"cs-projects-ink-eth-monitor/internal/metrics"
	"cs-projects-ink-eth-monitor/internal/monitor"
)

var (
	configPath = flag.String("config", "conf/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.Get()
	log.Info("启动监控服务", zap.String("config_path", *configPath))

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建客户端管理器
	clientManager, err := client.NewClientManager(cfg, log)
	if err != nil {
		log.Fatal("创建客户端管理器失败", zap.Error(err))
	}
	defer clientManager.Close()

	// 创建指标管理器
	metricsManager := metrics.NewMetrics(&cfg.Prometheus, log)
	defer metricsManager.Close()

	// 创建应急响应管理器
	emergencyManager, err := emergency.NewManager(&cfg.Emergency, cfg.InkRPC, log)
	if err != nil {
		log.Fatal("创建应急响应管理器失败", zap.Error(err))
	}
	defer emergencyManager.Close()

	// 创建监控器
	m := monitor.NewMonitor(cfg, clientManager, metricsManager, emergencyManager, log)

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动监控
	go func() {
		if err := m.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("监控服务异常", zap.Error(err))
			cancel()
		}
	}()

	// 等待信号
	select {
	case sig := <-sigChan:
		log.Info("收到退出信号", zap.String("signal", sig.String()))
	case <-ctx.Done():
		log.Info("监控服务已停止")
	}

	// 优雅关闭
	log.Info("退出...")

	// 停止监控
	m.Stop()

	// 最后推送一次指标
	if err := metricsManager.Push(); err != nil {
		log.Error("最终推送指标失败", zap.Error(err))
	}

	log.Info("监控服务已成功推出")
}
