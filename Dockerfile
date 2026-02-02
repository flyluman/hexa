FROM --platform=$BUILDPLATFORM golang:1.26-alpine3.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath \
    -ldflags="-s -w -buildid=" \
    -o /out/portfolio ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /out/portfolio /app/portfolio

ENV HTTP_PORT=8080
EXPOSE 8080

USER appuser:appgroup

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${HTTP_PORT}/health" > /dev/null || exit 1

ENTRYPOINT ["/app/portfolio"]
