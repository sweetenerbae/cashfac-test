FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /cashfac-news-api ./cmd/api

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /cashfac-news-api /usr/local/bin/cashfac-news-api

EXPOSE 8080

CMD ["cashfac-news-api"]
