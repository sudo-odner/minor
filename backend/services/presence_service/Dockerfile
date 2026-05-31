FROM golang:1.26.0-alpine AS builder

WORKDIR /src
COPY minor-shared /src/minor-shared
COPY minor/backend /src/minor/backend

WORKDIR /src/minor/backend/services/presence_service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/presence/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /src/minor/backend/services/presence_service/main .
EXPOSE 50086
CMD ["./main"]
