# 第一阶段：构建 Go 应用
FROM golang:1.24.6-alpine AS builder
# 配置 Go 模块和编译环境
ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 安装时区数据并设置上海时区
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY app/go.mod app/go.sum ./

# 下载依赖包
RUN go mod download

COPY app/ ./

RUN go build -v -o app ./cmd/api/main.go

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 设置运行时工作目录
WORKDIR /dist

COPY --from=builder /app/app ./app

# 可选：复制配置文件（如果应用需要读取本地配置，必须添加）
# COPY app/conf/ ./conf/

RUN chmod +x ./app

EXPOSE 8080

CMD ["./app"]