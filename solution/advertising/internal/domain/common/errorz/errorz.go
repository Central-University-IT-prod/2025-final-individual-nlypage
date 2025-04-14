package errorz

import "github.com/labstack/echo/v4"

var (
	ErrNotFound = &echo.HTTPError{
		Code:    echo.ErrNotFound.Code,
		Message: "Not Found",
	}
	ErrInternal = &echo.HTTPError{
		Code:    echo.ErrInternalServerError.Code,
		Message: "Internal Server Error",
	}
)
