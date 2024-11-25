package handler

import (
	"encoding/json"
	"net/http"

	"github.com/insomnius/wallet-event-loop/agregation"
	"github.com/labstack/echo/v4"
)

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserSigninRequest UserRegisterRequest

func UserRegister(authAggregator *agregation.Authorization) echo.HandlerFunc {
	return func(c echo.Context) error {
		var jsonBody UserRegisterRequest

		if err := json.NewDecoder(c.Request().Body).Decode((&jsonBody)); err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{
				"errors": []map[string]any{},
			})

			return err
		}

		if err := authAggregator.Register(jsonBody.Email, jsonBody.Password); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{
				"errors": []map[string]any{},
			})

			return err
		}

		return c.JSON(http.StatusOK, map[any]any{
			"data": jsonBody,
		})
	}
}

func UserSignin() echo.HandlerFunc {
	return func(c echo.Context) error {
		var jsonBody UserRegisterRequest

		if err := json.NewDecoder(c.Request().Body).Decode((&jsonBody)); err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{
				"errors": []map[string]any{},
			})

			return err
		}

		return nil
	}
}
