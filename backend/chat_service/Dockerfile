FROM golang:1.26.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main cmd/chat/main.go


FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8011

CMD ["./main"]


