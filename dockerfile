# ===== Build stage =====
FROM golang:1.25-alpine3.21 AS builder

WORKDIR /app
RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o url-shortener main.go

# ===== Runtime stage =====
FROM alpine:3.21

WORKDIR /app
RUN adduser -D appuser
USER appuser

COPY --from=builder /app/url-shortener /app/url-shortener

EXPOSE 8080
EXPOSE 9090

CMD ["./url-shortener"]
