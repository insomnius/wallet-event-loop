package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insomnius/wallet-event-loop/aggregation"
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
	"github.com/insomnius/wallet-event-loop/handler"
	"github.com/insomnius/wallet-event-loop/repository"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func setupTest() (*echo.Echo, *aggregation.Transaction, entity.User, entity.User, *db.Instance) {
	// Initialize the db instance and mock repositories
	dbInstance := db.NewInstance()

	dbInstance.CreateTable("users")
	dbInstance.CreateTable("user_tokens")
	dbInstance.CreateTable("wallets")
	dbInstance.CreateTable("transactions")
	dbInstance.CreateTable("mutations")

	userRepo := repository.NewUser(dbInstance)
	walletRepo := repository.NewWallet(dbInstance)
	mutationRepo := repository.NewMutation(dbInstance)

	// Create the transaction aggregator
	transactionAggregator := aggregation.NewTransaction(walletRepo, userRepo, mutationRepo, dbInstance)

	// Create users for testing
	user1 := entity.User{ID: "user-id-1", Email: "UserOne@test.com"}
	user2 := entity.User{ID: "user-id-2", Email: "UserTwo@test.com"}

	// Save users to the database
	userRepo.Put(user1)
	userRepo.Put(user2)

	// Create wallets for the users
	wallet1 := entity.Wallet{ID: "wallet-id-1", UserID: "user-id-1", Balance: 100}
	wallet2 := entity.Wallet{ID: "wallet-id-2", UserID: "user-id-2", Balance: 50}

	// Save wallets to the database
	walletRepo.Put(wallet1)
	walletRepo.Put(wallet2)

	// Create a new Echo instance
	e := echo.New()

	return e, transactionAggregator, user1, user2, dbInstance
}

func TestTopUp(t *testing.T) {
	e, trxAggregator, user1, _, _ := setupTest()

	// Mock the amount to top up
	amount := 100
	payload := map[string]int{"amount": amount}
	payloadBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/me/topup", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("current_user", entity.UserToken{
		UserID: user1.ID,
	})

	assert.NoError(t, handler.TopUp(trxAggregator)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "TopUp successful")
}

func TestCheckBalance(t *testing.T) {
	e, _, user1, _, dbInstance := setupTest()

	// Create a GET request to check user1's balance
	req := httptest.NewRequest(http.MethodGet, "/users/"+user1.ID+"/balance", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("current_user", entity.UserToken{
		UserID: user1.ID,
	})
	assert.NoError(t, handler.CheckBalance(repository.NewWallet(dbInstance))(c))

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "balance")
}

func TestTransfer(t *testing.T) {
	e, trxAggregator, user1, user2, _ := setupTest()

	// Create the transfer amount
	amount := 50
	payload := map[string]any{"amount": amount, "to": user2.ID}
	payloadBytes, _ := json.Marshal(payload)

	// Create a POST request to transfer funds from user1 to user2
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Set("current_user", entity.UserToken{
		UserID: user1.ID,
	})
	assert.NoError(t, handler.Transfer(trxAggregator)(c))

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Transfer successful")
}

func TestTopTransfer(t *testing.T) {
	e, trxAggregator, user1, user2, dbInstance := setupTest()

	// Create the transfer amount
	amount := 10
	payload := map[string]any{"amount": amount, "to": user2.ID}
	payloadBytes, _ := json.Marshal(payload)

	// Create a POST request to transfer funds from user1 to user2
	req := httptest.NewRequest(http.MethodPost, "/transactions/transfer", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Set("current_user", entity.UserToken{
		UserID: user1.ID,
	})
	assert.NoError(t, handler.Transfer(trxAggregator)(c))

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Transfer successful")

	// Create the transfer amount
	amount = 20
	payload = map[string]any{"amount": amount, "to": user2.ID}
	payloadBytes, _ = json.Marshal(payload)

	// Create a POST request to transfer funds from user1 to user2
	req = httptest.NewRequest(http.MethodPost, "/transactions/transfer", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	c = e.NewContext(req, rec)
	c.Set("current_user", entity.UserToken{
		UserID: user1.ID,
	})
	assert.NoError(t, handler.Transfer(trxAggregator)(c))

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Transfer successful")

	// Create the transfer amount
	amount = 30
	payload = map[string]any{"amount": amount, "to": user1.ID}
	payloadBytes, _ = json.Marshal(payload)

	// Create a POST request to transfer funds from user2 to user1
	req = httptest.NewRequest(http.MethodPost, "/transactions/transfer", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	c = e.NewContext(req, rec)
	c.Set("current_user", entity.UserToken{
		UserID: user2.ID,
	})
	assert.NoError(t, handler.Transfer(trxAggregator)(c))

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Transfer successful")

	// Create the transfer amount
	amount = 10
	payload = map[string]any{"amount": amount, "to": user1.ID}
	payloadBytes, _ = json.Marshal(payload)

	// Create a POST request to transfer funds from user2 to user1
	req = httptest.NewRequest(http.MethodPost, "/transactions/transfer", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	c = e.NewContext(req, rec)
	c.Set("current_user", entity.UserToken{
		UserID: user2.ID,
	})
	assert.NoError(t, handler.Transfer(trxAggregator)(c))

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Transfer successful")

	// Create a GET top transfer
	req = httptest.NewRequest(http.MethodGet, "/wallet/top-transfer", nil)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	// Call the handler
	c = e.NewContext(req, rec)
	c.Set("current_user", entity.UserToken{
		UserID: user1.ID,
	})
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.NoError(t, handler.TopTransfer(repository.NewMutation(dbInstance))(c))
	var jsonResponse map[string]map[string][]entity.Mutation

	err := json.NewDecoder(rec.Body).Decode(&jsonResponse)
	assert.NoError(t, err)

	assert.Equal(t, 30, jsonResponse["data"]["incoming"][0].Amount, jsonResponse)
	assert.Equal(t, 10, jsonResponse["data"]["incoming"][1].Amount, jsonResponse)
	assert.Equal(t, 20, jsonResponse["data"]["outgoing"][0].Amount, jsonResponse)
	assert.Equal(t, 10, jsonResponse["data"]["outgoing"][1].Amount, jsonResponse)
}
