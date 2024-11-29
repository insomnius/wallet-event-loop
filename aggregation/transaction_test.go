package aggregation_test

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/insomnius/wallet-event-loop/aggregation"
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
	"github.com/insomnius/wallet-event-loop/repository"
	"github.com/stretchr/testify/assert"
)

func setupDB() *db.Instance {
	dbInstance := db.NewInstance()

	dbInstance.CreateTable("users")
	dbInstance.CreateTable("wallets")
	dbInstance.CreateTable("mutations")
	return dbInstance
}

func TestTopUp(t *testing.T) {
	dbInstance := setupDB()

	// Repositories
	userRepo := repository.NewUser(dbInstance)
	walletRepo := repository.NewWallet(dbInstance)
	mutationRepo := repository.NewMutation(dbInstance)

	// Add user and wallet
	userID := uuid.New().String()
	err := userRepo.Put(entity.User{ID: userID, Email: "test@example.com"})
	assert.NoError(t, err)

	walletID := uuid.New().String()
	err = walletRepo.Put(entity.Wallet{ID: walletID, UserID: userID, Balance: 0})
	assert.NoError(t, err)

	// Transaction service
	transaction := aggregation.NewTransaction(walletRepo, userRepo, mutationRepo, dbInstance)

	// Case: Successful top-up
	err = transaction.TopUp(userID, 100)
	assert.NoError(t, err)

	// Verify wallet balance
	wallet, err := walletRepo.FindByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, 100, wallet.Balance)

	// Verify mutation
	mutations, err := mutationRepo.GetByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mutations))
	assert.Equal(t, 100, mutations[0].Amount)
	assert.Equal(t, entity.MutationTypeCredit, mutations[0].Type) // 1 for credit

	// Case: Non-existent user
	err = transaction.TopUp("non-existent-user", 50)
	assert.Error(t, err)

	// Case: Non-existent wallet
	userWithoutWallet := uuid.New().String()
	err = userRepo.Put(entity.User{ID: userWithoutWallet, Email: "nowallet@example.com"})
	assert.NoError(t, err)
	err = transaction.TopUp(userWithoutWallet, 50)
	assert.Error(t, err)
}

func TestTransfer(t *testing.T) {
	dbInstance := setupDB()

	// Repositories
	userRepo := repository.NewUser(dbInstance)
	walletRepo := repository.NewWallet(dbInstance)
	mutationRepo := repository.NewMutation(dbInstance)

	// Add users and wallets
	sourceUserID := uuid.New().String()
	targetUserID := uuid.New().String()

	err := userRepo.Put(entity.User{ID: sourceUserID, Email: "source@example.com"})
	assert.NoError(t, err)
	err = userRepo.Put(entity.User{ID: targetUserID, Email: "target@example.com"})
	assert.NoError(t, err)

	sourceWalletID := uuid.New().String()
	targetWalletID := uuid.New().String()

	err = walletRepo.Put(entity.Wallet{ID: sourceWalletID, UserID: sourceUserID, Balance: 200})
	assert.NoError(t, err)
	err = walletRepo.Put(entity.Wallet{ID: targetWalletID, UserID: targetUserID, Balance: 50})
	assert.NoError(t, err)

	// Transaction service
	transaction := aggregation.NewTransaction(walletRepo, userRepo, mutationRepo, dbInstance)

	// Case: Successful transfer
	err = transaction.Transfer(sourceUserID, targetUserID, 100)
	assert.NoError(t, err)

	// Verify balances
	sourceWallet, err := walletRepo.FindByUserID(sourceUserID)
	assert.NoError(t, err)
	targetWallet, err := walletRepo.FindByUserID(targetUserID)
	assert.NoError(t, err)
	assert.Equal(t, 100, sourceWallet.Balance)
	assert.Equal(t, 150, targetWallet.Balance)

	// Verify mutations
	sourceMutations, err := mutationRepo.GetByUserID(sourceUserID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(sourceMutations))
	assert.Equal(t, 100, sourceMutations[0].Amount)
	assert.Equal(t, entity.MutationTypeDebit, sourceMutations[0].Type) // 0 for debit

	targetMutations, err := mutationRepo.GetByUserID(targetUserID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(targetMutations))
	assert.Equal(t, 100, targetMutations[0].Amount)
	assert.Equal(t, entity.MutationTypeCredit, targetMutations[0].Type) // 1 for credit

	// Case: Insufficient funds
	err = transaction.Transfer(sourceUserID, targetUserID, 300)
	assert.ErrorIs(t, err, aggregation.ErrInsuficientFound)

	// Case: Non-existent user
	err = transaction.Transfer("non-existent-user", targetUserID, 50)
	assert.Error(t, err)

	// Case: Non-existent wallet
	userWithoutWallet := uuid.New().String()
	err = userRepo.Put(entity.User{ID: userWithoutWallet, Email: "nowallet@example.com"})
	assert.NoError(t, err)
	err = transaction.Transfer(userWithoutWallet, targetUserID, 50)
	assert.Error(t, err)
}

func TestRaceCondition(t *testing.T) {
	// Initialize the in-memory database instance
	dbInstance := db.NewInstance()

	// Create necessary tables
	dbInstance.CreateTable("users")
	dbInstance.CreateTable("wallets")
	dbInstance.CreateTable("mutations")

	// Set up repositories
	userRepo := repository.NewUser(dbInstance)
	walletRepo := repository.NewWallet(dbInstance)
	mutationRepo := repository.NewMutation(dbInstance)

	// Set up transaction aggregator
	transactionAggregator := aggregation.NewTransaction(walletRepo, userRepo, mutationRepo, dbInstance)

	// Initialize test data
	userID := uuid.New().String()
	targetID := uuid.New().String()

	userRepo.Put(entity.User{ID: userID, Email: "user1@example.com"})
	userRepo.Put(entity.User{ID: targetID, Email: "target@example.com"})

	sourceWallet := entity.Wallet{ID: uuid.New().String(), UserID: userID, Balance: 1000}
	targetWallet := entity.Wallet{ID: uuid.New().String(), UserID: targetID, Balance: 500}

	walletRepo.Put(sourceWallet)
	walletRepo.Put(targetWallet)

	// Define concurrent operations
	var wg sync.WaitGroup
	routines := 10
	topUpAmount := 100
	transferAmount := 50

	// Concurrent TopUps
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := transactionAggregator.TopUp(userID, topUpAmount)
			assert.NoError(t, err)
		}()
	}

	// Concurrent Transfers
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := transactionAggregator.Transfer(userID, targetID, transferAmount)
			if err != nil && err != aggregation.ErrInsuficientFound {
				assert.NoError(t, err)
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Assert final balances
	finalSourceWallet, _ := walletRepo.FindByUserID(userID)
	finalTargetWallet, _ := walletRepo.FindByUserID(targetID)

	expectedSourceBalance := sourceWallet.Balance + (routines * topUpAmount) - (routines * transferAmount)
	expectedTargetBalance := targetWallet.Balance + (routines * transferAmount)

	assert.Equal(t, expectedSourceBalance, finalSourceWallet.Balance, "Source wallet balance mismatch")
	assert.Equal(t, expectedTargetBalance, finalTargetWallet.Balance, "Target wallet balance mismatch")
}

func BenchmarkTransfer(b *testing.B) {
	// Initialize the in-memory database instance
	dbInstance := db.NewInstance()

	// Create necessary tables
	dbInstance.CreateTable("users")
	dbInstance.CreateTable("wallets")
	dbInstance.CreateTable("mutations")

	// Set up repositories
	userRepo := repository.NewUser(dbInstance)
	walletRepo := repository.NewWallet(dbInstance)
	mutationRepo := repository.NewMutation(dbInstance)

	// Set up transaction aggregator
	transactionAggregator := aggregation.NewTransaction(walletRepo, userRepo, mutationRepo, dbInstance)

	// Initialize test data
	sourceUserID := uuid.New().String()
	targetUserID := uuid.New().String()

	userRepo.Put(entity.User{ID: sourceUserID, Email: "source@example.com"})
	userRepo.Put(entity.User{ID: targetUserID, Email: "target@example.com"})

	sourceWallet := entity.Wallet{ID: uuid.New().String(), UserID: sourceUserID, Balance: 1000}
	targetWallet := entity.Wallet{ID: uuid.New().String(), UserID: targetUserID, Balance: 500}

	walletRepo.Put(sourceWallet)
	walletRepo.Put(targetWallet)

	// Reset the timer to exclude setup time
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := transactionAggregator.Transfer(sourceUserID, targetUserID, 100)
		if err != nil && err != aggregation.ErrInsuficientFound {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
