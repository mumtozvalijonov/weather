FROM golang:1.26.3-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/weather-api ./cmd/api

FROM alpine:3.23

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/weather-api /app/weather-api

USER app
EXPOSE 8000

ENTRYPOINT ["/app/weather-api"]
