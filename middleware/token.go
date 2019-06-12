package middleware

import (
	"fmt"
	"github.com/labstack/echo"
)

func TokenHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")
		if token != "" {
			c.Request().Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		}
		return next(c)
	}
}
