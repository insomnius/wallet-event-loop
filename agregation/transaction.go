package agregation

import (
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/repository"
)

type Transaction struct {
	walletRepo   *repository.Wallet
	userRepo     *repository.User
	mutationRepo *repository.Mutation
	db           *db.Instance
}

func NewTransaction(walletRepo *repository.Wallet, userRepo *repository.User, mutationRepo *repository.Mutation, db *db.Instance) *Transaction {
	return &Transaction{
		walletRepo:   walletRepo,
		userRepo:     userRepo,
		mutationRepo: mutationRepo,
	}
}

func (t Transaction) TopUp() {

}

func (t Transaction) Transfer() {

}
