FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mock-courseware-platform .

FROM alpine:latest

# 安装ca-certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从builder阶段复制二进制文件
COPY --from=builder /app/mock-courseware-platform .

# 复制静态文件
COPY --from=builder /app/static ./static

# 创建数据目录
RUN mkdir -p data

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./mock-courseware-platform"] 