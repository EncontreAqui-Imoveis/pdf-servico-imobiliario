FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pdf-service ./cmd/server/main.go

FROM alpine:latest AS runner

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/pdf-service ./pdf-service

EXPOSE 8080

CMD ["./pdf-service"]
