package domain

import "fmt"

type ErrorCode string

var (
	NotFoundErrorCode           ErrorCode = "NOT_FOUND"
	UnauthorizedErrorCode       ErrorCode = "UNAUTHORIZED"
	ForbiddenErrorCode          ErrorCode = "FORBIDDEN"
	ValidationFailedErrorCode   ErrorCode = "VALIDATION_FAILED"
	DuplicateEntryErrorCode     ErrorCode = "DUPLICATE_ENTRY"
	ServerErrorCode             ErrorCode = "SERVER_ERROR"
	BusinessValidationErrorCode ErrorCode = "BUSINESS_VALIDATION_FAILED"
)

type DomainError struct {
	Message string
	Code    ErrorCode
	Meta    interface{}
	Cause   error
}

func (de DomainError) Error() string {
	return fmt.Sprintf("[%s] %v", de.Code, de.Message)
}

func (de DomainError) Unwrap() error {
	return de.Cause
}

func NotFoundError(message string) DomainError {
	err := DomainError{
		Message: message,
		Code:    NotFoundErrorCode,
	}

	return err
}

func UnauthorizedError(message string) DomainError {
	err := DomainError{
		Message: message,
		Code:    UnauthorizedErrorCode,
	}
	return err
}

func ForbiddenError(message string) DomainError {
	err := DomainError{
		Message: message,
		Code:    ForbiddenErrorCode,
	}
	return err
}

func ValidationFailedError(errors map[string][]string) DomainError {
	err := DomainError{
		Message: "Validation Failed",
		Code:    ValidationFailedErrorCode,
		Meta:    errors,
	}
	return err
}

func BusinessValidationError(message string) DomainError {
	err := DomainError{
		Message: message,
		Code:    BusinessValidationErrorCode,
	}
	return err
}

func DuplicateEntryError(message string) DomainError {
	err := DomainError{
		Message: message,
		Code:    DuplicateEntryErrorCode,
	}
	return err
}

func ServerError(message string, cause error) DomainError {
	err := DomainError{
		Message: message,
		Code:    ServerErrorCode,
		Cause:   cause,
	}
	return err
}
