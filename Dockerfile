FROM --platform=$BUILDPLATFORM node:16 AS builder

WORKDIR /web
COPY ./VERSION .
COPY ./web .

RUN npm config set registry https://registry.npmmirror.com && \
    npm install --prefix /web/default --legacy-peer-deps --retry 5 & \
    npm install --prefix /web/berry --legacy-peer-deps --retry 5 & \
    npm install --prefix /web/air --legacy-peer-deps --retry 5 & \
    wait

RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/default & \
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/berry & \
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/air & \
    wait

FROM golang:alpine AS builder2

RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev \
    build-base

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=off

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download -x

COPY . .
COPY --from=builder /web/build ./web/build

RUN go build -trimpath -ldflags "-s -w -X 'github.com/songquanpeng/one-api/common.Version=$(cat VERSION)' -linkmode external -extldflags '-static'" -o one-api

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder2 /build/one-api /

EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]