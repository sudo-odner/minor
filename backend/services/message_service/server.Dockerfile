FROM golang:1.26.0-alpine AS builder

WORKDIR /src
COPY minor-shared /src/minor-shared
COPY minor/backend /src/minor/backend

WORKDIR /src/minor/backend/services/message_service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/message/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /src/minor/backend/services/message_service/server .
EXPOSE 8083
CMD ["./server"]
