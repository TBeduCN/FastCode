# 第一阶段：构建二进制文件
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -ldflags "-X main.version=${VERSION:-dev} -X main.commit=$(git rev-parse --short HEAD)" -o fastcode .

# 第二阶段：创建纯净镜像
FROM alpine:latest

WORKDIR /app

# 复制二进制文件
COPY --from=builder /app/fastcode /app/fastcode

# 复制静态文件
COPY --from=builder /app/public /app/public

# 创建配置目录
RUN mkdir -p /app/config

# 暴露端口
EXPOSE 8080

# 运行命令
CMD ["./fastcode"]