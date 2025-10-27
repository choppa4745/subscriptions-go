FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org && go mod download
COPY . .
RUN go build -o /subscriptions ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /subscriptions /subscriptions
COPY .env /app/.env
WORKDIR /app
EXPOSE 8000
ENTRYPOINT ["/subscriptions"]
