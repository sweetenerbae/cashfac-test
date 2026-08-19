FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /cashfac-news-api ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /data \
    && chown app:app /data

COPY --from=builder /cashfac-news-api /usr/local/bin/cashfac-news-api

USER app

EXPOSE 8080

CMD ["cashfac-news-api"]
