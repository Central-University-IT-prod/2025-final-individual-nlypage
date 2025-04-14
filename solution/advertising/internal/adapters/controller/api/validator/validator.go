package validator

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"reflect"
	"strings"
	"unicode"
)

type Validator struct {
	validator *validator.Validate
}

type GlobalErrorHandlerResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       interface{}
}

func New() *Validator {
	newValidator := validator.New()

	_ = newValidator.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) >= 4 && len(fl.Field().String()) <= 20
	})

	_ = newValidator.RegisterValidation("code", func(fl validator.FieldLevel) bool {
		code := fl.Field().String()

		hasLength := len(code) == 6
		hasUppercase := strings.ToLower(code) != code
		hasDigit := strings.IndexFunc(code, func(c rune) bool { return unicode.IsDigit(c) }) != -1

		return hasLength && (hasUppercase || hasDigit)
	})

	_ = newValidator.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		hasMinLength := len(password) >= 8
		hasUppercase := strings.ToLower(password) != password
		hasLowercase := strings.ToUpper(password) != password
		hasDigit := strings.IndexFunc(password, func(c rune) bool { return unicode.IsDigit(c) }) != -1

		return hasMinLength && hasUppercase && hasLowercase && hasDigit
	})

	_ = newValidator.RegisterValidation("header", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) >= 5 && len(fl.Field().String()) <= 150
	})

	_ = newValidator.RegisterValidation("body", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) >= 5 && len(fl.Field().String()) <= 1500
	})

	return &Validator{
		newValidator,
	}
}

func (v Validator) ValidateData(data interface{}) *echo.HTTPError {
	var validationErrors []ErrorResponse

	// Handle nil input
	if data == nil {
		return &echo.HTTPError{
			Code:    echo.ErrBadRequest.Code,
			Message: "validation failed: input data is nil",
		}
	}

	// Check if data is a slice
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Slice {
		// Validate each element in the slice
		for i := 0; i < val.Len(); i++ {
			if err := v.validator.Struct(val.Index(i).Interface()); err != nil {
				var invalidValidationError *validator.InvalidValidationError
				if errors.As(err, &invalidValidationError) {
					return &echo.HTTPError{
						Code:    echo.ErrBadRequest.Code,
						Message: err.Error(),
					}
				}

				var validationErrs validator.ValidationErrors
				ok := errors.As(err, &validationErrs)
				if !ok {
					return &echo.HTTPError{
						Code:    echo.ErrBadRequest.Code,
						Message: "unexpected validation error type",
					}
				}

				for _, err := range validationErrs {
					var elem ErrorResponse
					elem.FailedField = fmt.Sprintf("[%d].%s", i, err.Field())
					elem.Tag = err.Tag()
					elem.Value = err.Value()
					elem.Error = true
					validationErrors = append(validationErrors, elem)
				}
			}
		}
	} else {
		// Validate single struct
		if err := v.validator.Struct(data); err != nil {
			var invalidValidationError *validator.InvalidValidationError
			if errors.As(err, &invalidValidationError) {
				return &echo.HTTPError{
					Code:    echo.ErrBadRequest.Code,
					Message: err.Error(),
				}
			}

			var validationErrs validator.ValidationErrors
			ok := errors.As(err, &validationErrs)
			if !ok {
				return &echo.HTTPError{
					Code:    echo.ErrBadRequest.Code,
					Message: "unexpected validation error type",
				}
			}

			for _, err := range validationErrs {
				var elem ErrorResponse
				elem.FailedField = err.Field()
				elem.Tag = err.Tag()
				elem.Value = err.Value()
				elem.Error = true
				validationErrors = append(validationErrors, elem)
			}
		}
	}

	if len(validationErrors) > 0 && validationErrors[0].Error {
		errMessages := make([]string, 0)
		for _, err := range validationErrors {
			errMessages = append(errMessages, fmt.Sprintf(
				"[%s]: '%v' | Needs to implement '%s'",
				err.FailedField,
				err.Value,
				err.Tag,
			))
		}

		return &echo.HTTPError{
			Code:    echo.ErrBadRequest.Code,
			Message: strings.Join(errMessages, " and "),
		}
	}
	return nil
}
