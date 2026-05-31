FROM golang:1.26.0-alpine AS builder

WORKDIR /src
COPY minor-shared /src/minor-shared
COPY minor/backend /src/minor/backend

WORKDIR /src/minor/backend/services/community_service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/community/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /src/minor/backend/services/community_service/main .
EXPOSE 8011
CMD ["./main"]
