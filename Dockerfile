# 构建阶段
FROM golang:1.24-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN make build

# 运行阶段
FROM alpine:latest

# 安装ca证书（用于HTTPS请求）
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1000 monitor && \
    adduser -D -u 1000 -G monitor monitor

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/monitor /app/monitor

# 创建配置目录
RUN mkdir -p /app/conf && chown -R monitor:monitor /app

# 切换到非root用户
USER monitor

# 设置入口点
ENTRYPOINT ["/app/monitor"]

# 默认参数
CMD ["-config=/app/conf/config.yaml"]
