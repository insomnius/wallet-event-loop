package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type H map[string]any

func renderInternalServerError(c echo.Context) {
	c.JSON(http.StatusInternalServerError, map[string]any{
		"errors": []map[string]any{
			{
				"detail": "internal server error",
			},
		},
	})
}
