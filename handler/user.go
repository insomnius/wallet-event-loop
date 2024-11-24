package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserSigninRequest UserRegisterRequest

func UserRegister(c echo.Context) error {
	var jsonBody UserRegisterRequest

	if err := json.NewDecoder(c.Request().Body).Decode((&jsonBody)); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{
			"errors": []map[string]any{},
		})

		return err
	}

	return nil
}

func UserSignin(c echo.Context) error {
	return nil
}
