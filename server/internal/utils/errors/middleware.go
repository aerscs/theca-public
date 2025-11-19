package errors

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware возвращает middleware для глобальной обработки ошибок
func ErrorHandlerMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Проверка на наличие ошибок после выполнения запроса
		if len(c.Errors) > 0 {
			// Получаем последнюю ошибку
			err := c.Errors.Last().Err

			// Логируем ошибку
			var customErr *Error
			if errors.As(err, &customErr) {
				log.Error("API error",
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"error_code", customErr.Code,
					"error_message", customErr.Message,
					"original_error", customErr.Err,
				)
			} else {
				log.Error("Unexpected API error",
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"error", err.Error(),
				)
			}

			// Отправляем ответ клиенту
			RespondWithError(c, err)

			// Прерываем обработку
			c.Abort()
		}
	}
}

// ErrorHandling позволяет добавить ошибку в контекст запроса и прервать обработку
func ErrorHandling(c *gin.Context, err error) {
	// Добавляем ошибку в контекст
	_ = c.Error(err)

	// Прерываем обработку
	c.Abort()
}
