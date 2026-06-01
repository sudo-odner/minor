FROM golang:1.26.0-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарный файл мигратора
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o migrator cmd/migrator/main.go

# Финальный этап
FROM alpine:latest

WORKDIR /root/

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/migrator .

# Запуск мигратора
CMD ["./migrator"]