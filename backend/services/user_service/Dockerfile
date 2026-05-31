FROM golang:1.26.0-alpine AS builder

WORKDIR /src
COPY minor-shared /src/minor-shared
COPY minor/backend /src/minor/backend

WORKDIR /src/minor/backend/services/user_service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/user/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /src/minor/backend/services/user_service/main .
EXPOSE 8084
CMD ["./main"]
