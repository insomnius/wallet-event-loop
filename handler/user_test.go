// This unit test generated by AI
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insomnius/wallet-event-loop/aggregation"
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/handler"
	"github.com/insomnius/wallet-event-loop/repository"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestUserHandlers(t *testing.T) {
	// Initialize database instance
	dbInstance := db.NewInstance()
	defer dbInstance.Close()

	// Start database instance
	go func() {
		dbInstance.Start()
	}()

	// Setup tables
	dbInstance.CreateTable("users")
	dbInstance.CreateTable("wallets")
	dbInstance.CreateTable("user_tokens")

	// Initialize repositories
	walletRepo := repository.NewWallet(dbInstance)
	userRepo := repository.NewUser(dbInstance)
	userTokenRepo := repository.NewUserToken(dbInstance)

	// Initialize aggregator
	authAggregator := aggregation.NewAuthorization(walletRepo, userRepo, userTokenRepo, dbInstance)

	// Initialize Echo server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Register handlers
	e.POST("/users", handler.UserRegister(authAggregator))
	e.POST("/users/signin", handler.UserSignin(authAggregator))

	t.Run("Register Success", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, handler.UserRegister(authAggregator)(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"email":"test@example.com"`)
		}
	})

	t.Run("Register Duplicate Email", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := handler.UserRegister(authAggregator)(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "user already registered")
	})

	t.Run("SignIn Success", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users/signin", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, handler.UserSignin(authAggregator)(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"token":`)
		}
	})

	t.Run("SignIn User Not Found", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "notfound@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users/signin", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := handler.UserSignin(authAggregator)(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "user not found")
	})

	t.Run("SignIn Authentication Failed", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users/signin", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := handler.UserSignin(authAggregator)(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "email and password doesn't match")
	})
}
