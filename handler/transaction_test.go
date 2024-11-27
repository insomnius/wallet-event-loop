package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insomnius/wallet-event-loop/agregation"
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
	"github.com/insomnius/wallet-event-loop/handler"
	"github.com/insomnius/wallet-event-loop/repository"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func setupTest() (*echo.Echo, *agregation.Transaction, entity.User, entity.User) {
	// Initialize the db instance and mock repositories
	dbInstance := db.NewInstance()

	// This is not effective, but I let it be for now
	// PR is open tho
	go dbInstance.Start()

	dbInstance.CreateTable("users")
	dbInstance.CreateTable("user_tokens")
	dbInstance.CreateTable("wallets")
	dbInstance.CreateTable("transactions")
	dbInstance.CreateTable("mutations")

	userRepo := repository.NewUser(dbInstance)
	walletRepo := repository.NewWallet(dbInstance)
	mutationRepo := repository.NewMutation(dbInstance)

	// Create the transaction aggregator
	transactionAggregator := agregation.NewTransaction(walletRepo, userRepo, mutationRepo, dbInstance)

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

	// Register the handlers
	e.POST("/users/:user_id/topup", handler.TopUp(transactionAggregator))
	e.GET("/users/:user_id/balance", handler.CheckBalance(walletRepo))
	e.POST("/users/:user_id/transfer/:target_id", handler.Transfer(transactionAggregator))

	return e, transactionAggregator, user1, user2
}

func TestTopUp(t *testing.T) {
	e, _, user1, _ := setupTest()

	// Mock the amount to top up
	amount := 100
	payload := map[string]int{"amount": amount}
	payloadBytes, _ := json.Marshal(payload)

	// Create a POST request to top-up user1 balance
	req := httptest.NewRequest(http.MethodPost, "/users/"+user1.ID+"/topup", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Call the handler
	e.ServeHTTP(rec, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "TopUp successful")
}

func TestCheckBalance(t *testing.T) {
	e, _, user1, _ := setupTest()

	// Create a GET request to check user1's balance
	req := httptest.NewRequest(http.MethodGet, "/users/"+user1.ID+"/balance", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	e.ServeHTTP(rec, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "balance")
}

func TestTransfer(t *testing.T) {
	e, _, user1, user2 := setupTest()

	// Create the transfer amount
	amount := 50
	payload := map[string]int{"amount": amount}
	payloadBytes, _ := json.Marshal(payload)

	// Create a POST request to transfer funds from user1 to user2
	req := httptest.NewRequest(http.MethodPost, "/users/"+user1.ID+"/transfer/"+user2.ID, bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Call the handler
	e.ServeHTTP(rec, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Transfer successful")
}
