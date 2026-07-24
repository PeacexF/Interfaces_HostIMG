FROM golang:1.26.2-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder /server /app/server

USER app

EXPOSE 8081
CMD ["/app/server"]
