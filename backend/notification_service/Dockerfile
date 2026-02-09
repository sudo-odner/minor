FROM golang:1.25.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main cmd/notification/main.go


FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8013

CMD ["./main"]


