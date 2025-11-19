package errors

import (
	"errors"
	"net/http"
)

// ErrorCode represents the error code that will be returned to the client
type ErrorCode string

// Constants with error codes
const (
	// General error codes
	CodeUnknownError    ErrorCode = "UNKNOWN_ERROR"
	CodeInvalidRequest  ErrorCode = "INVALID_REQUEST"
	CodeInternalError   ErrorCode = "INTERNAL_ERROR"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	CodeForbidden       ErrorCode = "FORBIDDEN"
	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// User-specific error codes
	CodeUserNotFound              ErrorCode = "USER_NOT_FOUND"
	CodeUserAlreadyExists         ErrorCode = "USER_ALREADY_EXISTS"
	CodeUserUsernameAlreadyExists ErrorCode = "USER_USERNAME_ALREADY_EXISTS"
	CodeUserEmailAlreadyExists    ErrorCode = "USER_EMAIL_ALREADY_EXISTS"
	CodeInvalidPassword           ErrorCode = "INVALID_PASSWORD"
	CodeInvalidEmail              ErrorCode = "INVALID_EMAIL"
	CodeInvalidUsername           ErrorCode = "INVALID_USERNAME"
	CodeInvalidVerificationCode   ErrorCode = "INVALID_VERIFICATION_CODE"
	CodeInvalidRefreshToken       ErrorCode = "INVALID_REFRESH_TOKEN"

	// Error codes for data operations
	CodeDataNotFound ErrorCode = "DATA_NOT_FOUND"
	CodeDataInvalid  ErrorCode = "DATA_INVALID"
	CodeDataConflict ErrorCode = "DATA_CONFLICT"
)

// HTTPStatusMapping maps error codes to HTTP statuses
var HTTPStatusMapping = map[ErrorCode]int{
	// General codes
	CodeUnknownError:    http.StatusInternalServerError,
	CodeInvalidRequest:  http.StatusBadRequest,
	CodeInternalError:   http.StatusInternalServerError,
	CodeNotFound:        http.StatusNotFound,
	CodeUnauthorized:    http.StatusUnauthorized,
	CodeForbidden:       http.StatusForbidden,
	CodeTooManyRequests: http.StatusTooManyRequests,

	// User-specific codes
	CodeUserNotFound:              http.StatusNotFound,
	CodeUserAlreadyExists:         http.StatusConflict,
	CodeUserUsernameAlreadyExists: http.StatusConflict,
	CodeUserEmailAlreadyExists:    http.StatusConflict,
	CodeInvalidPassword:           http.StatusUnauthorized,
	CodeInvalidEmail:              http.StatusBadRequest,
	CodeInvalidUsername:           http.StatusBadRequest,
	CodeInvalidVerificationCode:   http.StatusBadRequest,
	CodeInvalidRefreshToken:       http.StatusUnauthorized,

	// Error codes for data operations
	CodeDataNotFound: http.StatusNotFound,
	CodeDataInvalid:  http.StatusBadRequest,
	CodeDataConflict: http.StatusConflict,
}

// APIError represents the error structure for API responses
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// Error represents an error with additional context for API
type Error struct {
	Err     error
	Code    ErrorCode
	Message string
}

// New creates a new error with the specified code and message
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewWithError creates a new error based on an existing error
func NewWithError(err error, code ErrorCode, message string) *Error {
	return &Error{
		Err:     err,
		Code:    code,
		Message: message,
	}
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// ToAPIError converts Error to APIError for response to client
func (e *Error) ToAPIError() APIError {
	return APIError{
		Code:    e.Code,
		Message: e.Message,
	}
}

// GetHTTPStatus returns the HTTP status for the error
func (e *Error) GetHTTPStatus() int {
	if status, ok := HTTPStatusMapping[e.Code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// IsErrorCode checks if the error corresponds to the specified code
func IsErrorCode(err error, code ErrorCode) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}
