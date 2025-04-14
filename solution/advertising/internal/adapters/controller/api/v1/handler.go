package v1

import "github.com/labstack/echo/v4"

type Handler interface {
	Setup(group *echo.Group)
}
