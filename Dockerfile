FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.io,direct

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bot main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bot .

COPY --from=builder /app/migrations ./migrations

CMD ["./bot"]








