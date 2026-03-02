# 第一阶段：构建 Go 应用
FROM golang:1.24.6-alpine AS builder
ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目代码（根目录下的所有文件）
COPY . ./

# 编译（注意：你的 main.go 在 cmd 目录下）
RUN go build -v -o app ./cmd/api/main.go

# 第二阶段：运行镜像
FROM alpine:latest
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /dist
COPY --from=builder /app/app ./app
RUN chmod +x ./app
EXPOSE 8080
CMD ["./app"]