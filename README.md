# Chain Monitor - Ethereum & INK 链监控服务

这是一个用于监控Ethereum和INK链上智能合约状态的服务，支持自动轮询合约方法并将指标推送到Prometheus Gateway。

## 功能特性

### 监控能力
- **Ethereum链监控**
  - 监控合约暂停状态 (`paused()`, `pause(address)`)
  - 监控价格源 (`latestAnswer()`)

- **INK链监控**
  - 监控合约暂停状态 (`getPaused(address)`)
  - 监控价格源并与Ethereum价格对比，超过阈值告警
  - 监控储备上限与总供应量差异，超过阈值告警

### 核心特性
- 单一可执行程序同时监控两条链
- 定时轮询合约状态（可配置间隔）
- 自动失败重试机制
- 指标推送到Prometheus Gateway
- 优雅关闭支持
- 结构化日志（使用zap）
- YAML配置文件

## 项目结构

```
.
├── cmd/
│   └── monitor/
│       └── main.go           # 主程序入口
├── internal/
│   ├── client/
│   │   └── client.go         # RPC客户端
│   ├── config/
│   │   └── config.go         # 配置加载
│   ├── logger/
│   │   └── logger.go         # 日志初始化
│   ├── metrics/
│   │   └── metrics.go        # Prometheus指标
│   └── monitor/
│       └── monitor.go        # 监控核心逻辑
├── pkg/
│   └── retry/
│       └── retry.go          # 重试机制
├── conf/
│   └── config.yaml           # 配置文件
├── Makefile                  # 构建脚本
└── README.md                 # 项目文档
```

## 快速开始

### 前置要求

- Go 1.24+
- Prometheus Gateway (可选，用于接收指标)

### 安装步骤

```bash
# 1. 创建配置文件
cp conf/config.example.yaml conf/config.yaml

# 2. 安装依赖
make install
```

### 配置

编辑 `conf/config.yaml` 文件，配置RPC节点和监控参数：

```yaml
# 日志配置
log:
  level: info
  format: json
  output: stdout

# Prometheus配置
prometheus:
  gateway_url: "http://localhost:9091"
  job_name: "chain_monitor"
  push_interval: 30

# 监控配置
monitor:
  poll_interval: 30
  retry_times: 3
  retry_delay: 5

# Ethereum配置
ethereum:
  rpc_url: "https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  contracts:
    - address: "0x95703e0982140D16f8ebA6d158FccEde42f04a4C"
      name: "contract1"
      type: "pause_with_identifier"
      method: "paused"
      method_params:
        - "0x0000000000000000000000000000000000000000"

# INK链配置
ink:
  rpc_url: "https://rpc-gel.inkonchain.com"
  contracts:
    - address: "0x96086C25d13943C80Ff9a19791a40Df6aFC08328"
      name: "ink_pause_checker"
      type: "get_paused"
      method: "getPaused"
      method_params:
        - "0x0000000000000000000000000000000000000000"
```

### 编译

```bash
make build
```

编译后的可执行文件位于 `bin/monitor`

### 运行

```bash
make run
```

或直接运行：

```bash
./bin/monitor -config=conf/config.yaml
```

## 使用说明

### 监控指标

所有指标会自动推送到Prometheus Gateway，指标名称格式：

```
chain_monitor_{chain}_{contract_name}_{type}
```

示例：
- `chain_monitor_ethereum_contract1_pause_with_identifier` - Ethereum合约1的暂停状态
- `chain_monitor_ink_ink_pause_checker_get_paused` - INK链暂停检查器状态

指标值说明：
- `0` - 未暂停 / false
- `1` - 已暂停 / true
- 对于价格源，值为实际价格

### 告警配置

#### 价格差异告警

监控INK价格与Ethereum价格的差异：

```yaml
ink:
  contracts:
    - address: "0x163131609562E578754aF12E998635BfCa56712C"
      name: "ink_price_feed"
      type: "price_feed"
      method: "latestAnswer"
      alert:
        type: "price_diff"
        threshold: 0.05          # 5%差异阈值
        compare_with: "eth_price_feed"
```

当 `abs(price1 - price2) / price1 > 0.05` 时，会记录警告日志。

#### 供应量差异告警

监控储备上限与总供应量的差异：

```yaml
ink:
  contracts:
    - address: "0x96086C25d13943C80Ff9a19791a40Df6aFC08328"
      name: "ink_reserve_cap"
      type: "reserve_cap"
      method: "getReserveCap"
      method_params:
        - "0x4200000000000000000000000000000000000006"
      alert:
        type: "supply_diff"
        threshold: "2500000000000000000000"  # 2500 * 10^18
        compare_address: "0xc1457AcfBaD2332b07B7651A4Da3176E8F3Bc9E4"
        compare_method: "totalSupply"
```

当 `totalSupply - reserveCap > threshold` 时，会记录警告日志。

## Docker 部署

### 使用 Docker Compose

```bash
# 编辑配置文件，填入远程Prometheus Gateway地址
vim conf/config.yaml

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f monitor

# 停止服务
docker-compose down
```

### 仅构建 Docker 镜像

```bash
make docker-build
```

或手动构建：

```bash
docker build -t chain-monitor:latest .
```

运行容器：

```bash
docker run -d \
  --name chain-monitor \
  -p 8080:8080 \
  -v $(pwd)/conf/config.yaml:/app/conf/config.yaml:ro \
  chain-monitor:latest
```

## Makefile 命令

```bash
make build        # 编译项目
make run          # 编译并运行
make clean        # 清理构建产物
make test         # 运行测试
make coverage     # 生成测试覆盖率报告
make lint         # 代码检查
make fmt          # 格式化代码
make install      # 安装依赖
make docker-build # 构建Docker镜像
make help         # 显示帮助
```

## 开发指南

### 添加新的合约监控

1. 在配置文件中添加合约配置
2. 如果需要新的合约类型，在 `internal/monitor/monitor.go` 中添加处理逻辑
3. 注册对应的指标

### 日志级别

支持的日志级别：`debug`, `info`, `warn`, `error`

在配置文件中修改：
```yaml
log:
  level: debug
```

### 重试机制

所有RPC调用都自动支持重试，可在配置文件中调整：

```yaml
monitor:
  retry_times: 3      # 重试次数
  retry_delay: 5      # 重试延迟（秒）
```

## 架构设计

### 监控流程

1. 启动时加载配置并初始化所有组件
2. 注册所有需要监控的指标
3. 开始定时轮询：
   - 并发调用所有合约方法
   - 处理返回结果并更新指标
   - 检查告警条件
   - 推送指标到Prometheus Gateway
5. 优雅关闭时推送最后一次指标

### 为什么选择轮询而非事件监听？

1. **统一性** - 有些监控任务（如价格对比、供应量检查）本身就需要轮询
2. **简单可靠** - 轮询逻辑简单，不需要维护WebSocket连接
3. **容错性** - RPC节点故障时自动重试，不会丢失监控
4. **成本** - 对于状态不频繁变化的场景，轮询开销可接受

### 为什么单一程序监控两条链？

1. **代码复用** - 共享客户端、重试、日志等基础设施
2. **配置统一** - 单一配置文件，易于管理
3. **部署简单** - 只需部署一个服务
4. **成本低** - 减少运维复杂度

## 故障排查

### 常见问题

**Q: 连接RPC节点失败**
```
A: 检查网络连接和RPC URL配置，确认API KEY是否正确
```

**Q: 推送Prometheus Gateway失败**
```
A: 检查Gateway是否运行，URL是否正确
```

**Q: 合约调用返回错误**
```
A: 检查合约地址和方法签名是否正确，查看日志中的详细错误信息
```

### 日志查看

服务使用结构化日志，可以通过以下方式查看：

```bash
# 控制台输出（开发环境）
./bin/monitor -config=conf/config.yaml

# 输出到文件（生产环境）
# 修改config.yaml中的log.output配置
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！
