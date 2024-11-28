package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/insomnius/wallet-event-loop/aggregation"
	"github.com/labstack/echo/v4"
)

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserSigninRequest UserRegisterRequest

func UserRegister(authAggregator *aggregation.Authorization) echo.HandlerFunc {
	return func(c echo.Context) error {
		var jsonBody UserRegisterRequest

		if err := json.NewDecoder(c.Request().Body).Decode((&jsonBody)); err != nil {
			c.JSON(http.StatusBadRequest, H{
				"errors": []H{
					{
						"detail": "bad json request",
					},
				},
			})

			return err
		}

		if err := authAggregator.Register(jsonBody.Email, jsonBody.Password); err != nil {
			if errors.Is(err, aggregation.ErrUserAlreadyExists) {
				// Could lead to security issue, but doesnt matter for now
				c.JSON(http.StatusUnprocessableEntity, H{
					"errors": []H{
						{
							"detail": "user already registered",
						},
					},
				})
				return err
			}

			renderInternalServerError(c)

			return err
		}

		return c.JSON(http.StatusOK, H{
			"data": jsonBody,
		})
	}
}

func UserSignin(authAggregator *aggregation.Authorization) echo.HandlerFunc {
	return func(c echo.Context) error {
		var jsonBody UserRegisterRequest

		if err := json.NewDecoder(c.Request().Body).Decode((&jsonBody)); err != nil {
			c.JSON(http.StatusBadRequest, H{
				"errors": []H{
					{
						"detail": "bad json request",
					},
				},
			})

			return err
		}

		token, err := authAggregator.SignIn(jsonBody.Email, jsonBody.Password)
		if err != nil {
			if errors.Is(err, aggregation.ErrUserNotFound) {
				// Could lead to security issue, but doesnt matter for now
				c.JSON(http.StatusNotFound, H{
					"errors": []H{
						{
							"detail": "user not found",
						},
					},
				})
				return err
			}

			if errors.Is(err, aggregation.ErrAuthFailed) {
				c.JSON(http.StatusUnprocessableEntity, H{
					"errors": []H{
						{
							"detail": "email and password doesn't match",
						},
					},
				})
				return err
			}

			renderInternalServerError(c)

			return err
		}

		return c.JSON(http.StatusOK, H{
			"data": H{
				"token": token,
			},
		})
	}
}
