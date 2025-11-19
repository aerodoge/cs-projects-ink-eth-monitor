package retry

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Func 重试函数类型
type Func func() error

// Do 执行带重试的函数
func Do(ctx context.Context, fn Func, maxRetries int, delay time.Duration, logger *zap.Logger) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 执行函数
		err = fn()
		if err == nil {
			return nil
		}

		// 如果是最后一次尝试，直接返回错误
		if i == maxRetries {
			break
		}

		// 记录重试日志
		logger.Warn("操作失败，准备重试",
			zap.Int("retry_count", i+1),
			zap.Int("max_retries", maxRetries),
			zap.Duration("delay", delay),
			zap.Error(err),
		)

		// 等待后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("重试%d次后仍然失败: %w", maxRetries, err)
}
