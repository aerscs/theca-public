package errors

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// FromGormError converts GORM errors into custom errors
func FromGormError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return New(CodeDataNotFound, "Record not found")
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return New(CodeDataConflict, "Record with this data already exists")
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return New(CodeDataInvalid, "Foreign key constraint violation")
	case errors.Is(err, gorm.ErrInvalidField):
		return New(CodeInvalidRequest, "Invalid data specified")
	default:
		// Check if there are keywords in the error text
		errMsg := strings.ToLower(err.Error())

		if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique") {
			return New(CodeDataConflict, "Record with this data already exists")
		}

		if strings.Contains(errMsg, "foreign key") {
			return New(CodeDataInvalid, "Foreign key constraint violation")
		}

		return NewWithError(err, CodeUnknownError, "Database error")
	}
}

// FromUserError converts database errors into specific user errors
// by analyzing the error message for user-related constraints
func FromUserError(err error, table string) error {
	if err == nil {
		return nil
	}

	// First try standard GORM error conversion
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return New(CodeUserNotFound, "User not found")
	}

	errMsg := strings.ToLower(err.Error())

	// Handle duplicate key violations specifically for users table
	if table == "users" && (strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique")) {
		if strings.Contains(errMsg, "email") || strings.Contains(errMsg, "idx_users_email") {
			return New(CodeUserEmailAlreadyExists, "User with this email already exists")
		}
		if strings.Contains(errMsg, "username") || strings.Contains(errMsg, "idx_users_username") {
			return New(CodeUserUsernameAlreadyExists, "User with this username already exists")
		}
		// Fallback to general user exists error
		return New(CodeUserAlreadyExists, "User already exists")
	}

	// Use the general converter for other errors
	return FromGormError(err)
}
