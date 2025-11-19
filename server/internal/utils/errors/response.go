package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents the standard API response format
type Response struct {
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
	Success bool      `json:"success"`
}

// SuccessResponse creates a successful response with data
func SuccessResponse(data any) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// ErrorResponse creates a response with an error
func ErrorResponse(err error) Response {
	var customErr *Error
	var apiErr APIError

	if errors.As(err, &customErr) {
		// If this is our custom error, use information from it
		apiErr = customErr.ToAPIError()
	} else {
		// If this is a standard error, use a general error code
		apiErr = APIError{
			Code:    CodeUnknownError,
			Message: err.Error(),
		}
	}

	return Response{
		Success: false,
		Error:   &apiErr,
	}
}

// RespondWithError sends a response with an error through gin.Context
func RespondWithError(c *gin.Context, err error) {
	var customErr *Error
	var statusCode int
	var response Response

	if errors.As(err, &customErr) {
		// For a custom error, use the corresponding HTTP status
		statusCode = customErr.GetHTTPStatus()
		response = ErrorResponse(err)
	} else {
		// For a standard error, use Internal Server Error
		statusCode = http.StatusInternalServerError
		response = ErrorResponse(err)
	}

	c.JSON(statusCode, response)
}

// RespondWithSuccess sends a successful response through gin.Context
func RespondWithSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, SuccessResponse(data))
}
