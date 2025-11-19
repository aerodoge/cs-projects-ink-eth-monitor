#!/bin/bash

# 合约集成测试脚本
# 使用真实的 RPC 端点测试所有合约

set -e

# 从配置文件读取 RPC URLs
CONFIG_FILE="conf/config.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 找不到配置文件 $CONFIG_FILE"
    exit 1
fi

# 提取 RPC URLs（简单的 grep 和 awk）
ETH_RPC=$(grep "^eth_rpc:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')
INK_RPC=$(grep "^ink_rpc:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')

if [ -z "$ETH_RPC" ] || [ -z "$INK_RPC" ]; then
    echo "错误: 无法从配置文件读取 RPC URLs"
    exit 1
fi

echo "================================================"
echo "运行合约集成测试"
echo "================================================"
echo "Ethereum RPC: $ETH_RPC"
echo "INK RPC: $INK_RPC"
echo "================================================"
echo ""

# 设置环境变量并运行测试
export TEST_ETH_RPC="$ETH_RPC"
export TEST_INK_RPC="$INK_RPC"

# 运行测试
cd "$(dirname "$0")/.."
go test -v ./internal/contracts -timeout 120s

echo ""
echo "================================================"
echo "测试完成"
echo "================================================"
