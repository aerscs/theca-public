# Build stage
FROM golang:1.24.2-alpine AS builder

# Устанавливаем необходимые инструменты для сборки
RUN apk add --no-cache git make

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
COPY vendor ./vendor

# Копируем исходный код
COPY . .

# Устанавливаем флаги сборки
ENV GOFLAGS="-mod=vendor" \
    CGO_ENABLED=0 \
    GOOS=linux

# Собираем приложение
RUN go build -o theca ./cmd/theca/main.go

# Final stage
FROM alpine:3.19

# Устанавливаем необходимые пакеты
RUN apk add --no-cache ca-certificates tzdata curl

# Создаем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/theca .
# Копируем шаблоны
COPY --from=builder /app/templates ./templates/

# Копируем непривилегированного пользователя
COPY --from=builder /etc/passwd /etc/passwd

# Добавляем метаданные
LABEL maintainer="Theca Team" \
    description="Theca API Server" \
    version="1.0.0"

# Добавляем healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/v1/health || exit 1

# Открываем порты
EXPOSE 8080

# Запускаем приложение
CMD ["./theca"]
