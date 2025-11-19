.PHONY: all build run daemon stop clean test lint fmt help install

# 变量定义
BINARY_NAME=monitor
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/monitor/main.go
CONFIG_PATH=conf/config.yaml

# Go命令
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# 构建标志
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-v

all: clean build

## build: 编译项目
build:
	@echo "编译项目..."
	@mkdir -p bin
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "编译完成: $(BINARY_PATH)"

## run: 运行程序
run: build
	@echo "运行监控服务..."
	./$(BINARY_PATH) -config=$(CONFIG_PATH)

## daemon: 后台运行程序
daemon: build
	@echo "后台启动监控服务..."
	@nohup ./$(BINARY_PATH) -config=$(CONFIG_PATH) > monitor.log 2>&1 &
	@sleep 2
	@if pgrep -f $(BINARY_PATH) > /dev/null; then \
		echo "✅ 监控服务启动成功"; \
		echo "PID: $$(pgrep -f $(BINARY_PATH))"; \
		echo "查看日志: tail -f monitor.log"; \
	else \
		echo "❌ 监控服务启动失败"; \
		echo "查看错误日志: tail monitor.log"; \
		exit 1; \
	fi

## stop: 停止后台服务
stop:
	@echo "停止监控服务..."
	@pkill -f $(BINARY_PATH) 2>/dev/null && echo "监控服务已停止" || echo "监控服务未运行"
	@true

## clean: 清理构建产物
clean:
	@echo "清理构建产物..."
	@$(GOCLEAN)
	@rm -rf bin/
	@echo "清理完成"

## test: 运行测试
test:
	@echo "运行测试..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "测试完成"

## coverage: 查看测试覆盖率
coverage: test
	@echo "生成覆盖率报告..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

## lint: 代码检查
lint:
	@echo "运行代码检查..."
	@which golangci-lint > /dev/null || (echo "请先安装golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...
	@echo "代码检查完成"

## fmt: 格式化代码
fmt:
	@echo "格式化代码..."
	$(GOFMT) ./...
	@echo "格式化完成"

## install: 安装依赖
install:
	@echo "安装依赖..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "依赖安装完成"

## deps: 查看依赖
deps:
	@echo "项目依赖:"
	$(GOMOD) graph

## update: 更新依赖
update:
	@echo "更新依赖..."
	$(GOMOD) tidy
	@echo "依赖更新完成"

## docker-build: 构建Docker镜像
docker-build:
	@echo "构建Docker镜像..."
	docker build -t chain-monitor:latest .
	@echo "Docker镜像构建完成"

## help: 显示帮助信息
help:
	@echo "可用命令:"
	@echo "  make build        - 编译项目"
	@echo "  make run          - 编译并运行程序"
	@echo "  make daemon       - 后台运行程序（日志输出到 monitor.log）"
	@echo "  make stop         - 停止后台服务"
	@echo "  make clean        - 清理构建产物"
	@echo "  make test         - 运行测试"
	@echo "  make coverage     - 生成测试覆盖率报告"
	@echo "  make lint         - 运行代码检查"
	@echo "  make fmt          - 格式化代码"
	@echo "  make install      - 安装依赖"
	@echo "  make deps         - 查看依赖"
	@echo "  make update       - 更新依赖"
	@echo "  make docker-build - 构建Docker镜像"
	@echo "  make help         - 显示此帮助信息"
