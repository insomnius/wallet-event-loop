package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// TokenValidator is a function type for validating tokens.
type TokenValidator func(c echo.Context, token string) (bool, error)

// Oauth creates an OAuth middleware to validate Bearer tokens.
func Oauth(validateToken TokenValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Retrieve the Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
			}

			// Check if the header has the Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization format")
			}

			token := parts[1]

			// Validate the token using the provided validator
			valid, err := validateToken(c, token)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Error validating token").SetInternal(err)
			}
			if !valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// Proceed to the next handler
			return next(c)
		}
	}
}
