# Theca - менеджер закладок

## О проекте

Theca - это современный менеджер закладок, разработанный для удобного хранения, организации и доступа к вашим веб-ссылкам. Название "Theca" происходит от латинского слова, обозначающего футляр или контейнер для хранения ценных предметов.

## Ключевые особенности

- **Эффективное управление закладками**: удобная организация и категоризация ссылок
- **API-архитектура**: полностью разделённые бэкенд и фронтенд части
- **Современные технологии**: Go, PostgreSQL, Redis
- **Аутентификация с JWT**: безопасная система авторизации
- **Документация Swagger**: интерактивная документация API
- **Элегантная обработка ошибок**: унифицированная система обработки и отображения ошибок

## Технологии

- **Go 1.24**: основной язык разработки
- **Gin**: веб-фреймворк для API
- **GORM**: ORM для работы с базами данных
- **PostgreSQL**: основная база данных
- **Redis**: кеширование и хранение токенов и кодов подтверждения почты
- **JWT**: авторизация и аутентификация
- **Docker & Docker Compose**: контейнеризация и оркестрация
- **Swagger**: документация API
- **Resend**: отправка email-уведомлений

## Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- Go 1.24+ (для локальной разработки)
- Make (опционально)

### Быстрый старт с Docker

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/OxytocinGroup/theca-v3.git
   cd theca-v3/server
   ```

2. Создайте файл .env с необходимыми переменными окружения:
   ```
   LOG_LEVEL=INFO
   PG_USER=postgres
   PG_PASSWORD=postgres
   PG_DB=postgres
   REDIS_PASSWORD=your_redis_password
   JWT_ACCESS_SECRET=your_jwt_access_secret
   JWT_REFRESH_SECRET=your_jwt_refresh_secret
   SMTP_API_KEY=your_smtp_api_key
   ```

3. Запустите проект через Docker Compose:
   ```bash
   docker-compose up -d
   ```

4. Сервер будет доступен по адресу: http://localhost:8080
   Swagger UI будет доступен по адресу: http://localhost:8081/swagger/index.html

### Локальная разработка

1. Установите зависимости:
   ```bash
   go mod download
   ```

2. Сгенерируйте документацию Swagger:
   ```bash
   make swag
   ```

3. Запустите локальный сервер:
   ```bash
   go run ./cmd/theca/main.go
   ```

## Система обработки ошибок в API

### Концепция

Система обработки ошибок в API построена на основе кодов ошибок, которые возвращаются клиенту. Каждый код ошибки соответствует определенной ситуации и помогает клиенту понять, что именно произошло.

### Структура ответа API

Все ответы API имеют следующую структуру:

```json
{
  "success": true|false,
  "data": {}, // только для успешных ответов
  "error": {  // только для ошибок
    "code": "ERROR_CODE",
    "message": "Сообщение об ошибке"
  }
}
```

### Коды ошибок

Коды ошибок разделены на несколько категорий:

#### Общие коды ошибок

- `UNKNOWN_ERROR` - неизвестная ошибка
- `INVALID_REQUEST` - неверный формат запроса
- `INTERNAL_ERROR` - внутренняя ошибка сервера
- `NOT_FOUND` - ресурс не найден
- `UNAUTHORIZED` - неавторизованный доступ
- `FORBIDDEN` - доступ запрещен

#### Пользовательские коды ошибок

- `USER_NOT_FOUND` - пользователь не найден
- `USER_ALREADY_EXISTS` - пользователь уже существует
- `INVALID_PASSWORD` - неверный пароль
- `INVALID_EMAIL` - неверный формат email
- `INVALID_USERNAME` - неверный формат имени пользователя
- `INVALID_VERIFICATION_CODE` - неверный код верификации
- `INVALID_REFRESH_TOKEN` - неверный токен обновления

#### Коды ошибок для операций с данными

- `DATA_NOT_FOUND` - данные не найдены
- `DATA_INVALID` - недопустимые данные
- `DATA_CONFLICT` - конфликт данных

### HTTP статусы

Каждый код ошибки соответствует определенному HTTP-статусу:

- **4XX Errors**
  - `INVALID_REQUEST` - 400 Bad Request
  - `INVALID_PASSWORD` - 400 Bad Request
  - `DATA_INVALID` - 400 Bad Request
  - `UNAUTHORIZED` - 401 Unauthorized
  - `FORBIDDEN` - 403 Forbidden
  - `NOT_FOUND` - 404 Not Found
  - `DATA_NOT_FOUND` - 404 Not Found
  - `USER_NOT_FOUND` - 404 Not Found
  - `USER_ALREADY_EXISTS` - 409 Conflict
  - `DATA_CONFLICT` - 409 Conflict

- **5XX Errors**
  - `UNKNOWN_ERROR` - 500 Internal Server Error

### Использование в коде
#### Создание новой ошибки
```go
// Создание новой ошибки
err := customerrors.New(customerrors.CodeUserNotFound, "Пользователь не найден")

// Создание ошибки на основе существующей
err := customerrors.NewWithError(originalError, customerrors.CodeInternalError, "Внутренняя ошибка сервера")
```

#### Отправка ответа с ошибкой

```go
// Отправка ответа с ошибкой
customerrors.RespondWithError(c, err)

// Отправка успешного ответа
customerrors.RespondWithSuccess(c, data)
```

#### Проверка типа ошибки

```go
// Проверка типа ошибки
if customerrors.IsErrorCode(err, customerrors.CodeUserNotFound) {
    // Обработка ошибки "Пользователь не найден"
}
```

## Документация API

API документация доступна через Swagger UI. После запуска сервера, документация будет доступна по адресу:

```
http://localhost:8081/swagger/index.html
```

Swagger предоставляет:
- Интерактивную документацию всех API endpoints
- Возможность тестирования API прямо из браузера
- Описание всех моделей данных и параметров запросов
- Информацию о кодах ошибок и их значениях 

## Архитектура проекта

Проект следует принципам чистой архитектуры и разделен на следующие основные компоненты:

- `cmd/theca`: точка входа в приложение
- `internal/app`: основная логика приложения
- `internal/config`: конфигурация приложения
- `internal/logger`: настройка логирования
- `internal/model`: модели данных
- `internal/repository`: слой доступа к данным
- `internal/service`: бизнес-логика
- `internal/server`: HTTP сервер и маршрутизация
- `internal/storage`: абстракции для работы с хранилищами
- `internal/utils`: вспомогательные утилиты, включая обработку ошибок

## Настройка rate limiting

Rate limiting автоматически включается при наличии Redis. Настройки:

- **Логин**: 5 попыток за 15 минут с одного IP
- **Сброс пароля**: 3 попытки за час с одного IP  
- **Верификация email**: 5 попыток отправки за 10 минут с одного IP
- **Ввод кода**: 5 попыток за час для каждого пользователя

## Лицензия

Apache License 2.0 