FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod tidy
COPY . .

RUN go build -o api ./cmd/api
RUN go build -o migrate ./cmd/migrate
RUN go build -o notification-service ./cmd/notification-service
RUN go build -o websocket-service ./cmd/websocket-service
RUN go build -o chat-service ./cmd/chat-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/migrate .
COPY --from=builder /app/notification-service .
COPY --from=builder /app/websocket-service .
COPY --from=builder /app/chat-service .
COPY --from=builder /app/migrations ./migrations

CMD ./migrate up && ./api
