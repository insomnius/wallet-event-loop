package agregation

import (
	"errors"

	"github.com/google/uuid"
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
	"github.com/insomnius/wallet-event-loop/repository"
)

type Transaction struct {
	walletRepo   *repository.Wallet
	userRepo     *repository.User
	mutationRepo *repository.Mutation
	db           *db.Instance
}

var ErrInsuficientFound = errors.New("error insuficient found")

func NewTransaction(
	walletRepo *repository.Wallet,
	userRepo *repository.User,
	mutationRepo *repository.Mutation,
	db *db.Instance,
) *Transaction {
	return &Transaction{
		walletRepo:   walletRepo,
		userRepo:     userRepo,
		mutationRepo: mutationRepo,
		db:           db,
	}
}

func (t Transaction) TopUp(userID string, amount int) error {
	return t.db.Transaction(func(trx *db.Transaction) error {
		user, err := t.userRepo.FindById(userID, trx)
		if err != nil {
			return err
		}

		wallet, err := t.walletRepo.FindByUserID(user.ID, trx)
		if err != nil {
			return err
		}

		wallet.Balance += amount
		if err := t.walletRepo.Put(wallet, trx); err != nil {
			return err
		}

		return t.mutationRepo.Put(entity.Mutation{
			ID:       uuid.New().String(),
			WalletID: wallet.ID,
			UserID:   userID,
			Type:     1, // topup
			Amount:   amount,
		}, trx)
	})
}

func (t Transaction) Transfer(userID, targetID string, amount int) error {
	return t.db.Transaction(func(trx *db.Transaction) error {
		user, err := t.userRepo.FindById(userID, trx)
		if err != nil {
			return err
		}

		target, err := t.userRepo.FindById(targetID, trx)
		if err != nil {
			return err
		}

		sourceWallet, err := t.walletRepo.FindByUserID(user.ID, trx)
		if err != nil {
			return err
		}

		if sourceWallet.Balance-amount < 0 {
			return ErrInsuficientFound
		}

		targetWallet, err := t.walletRepo.FindByUserID(target.ID, trx)
		if err != nil {
			return err
		}

		targetWallet.Balance += amount
		sourceWallet.Balance -= amount

		if err := t.walletRepo.Put(targetWallet, trx); err != nil {
			return err
		}

		if err := t.walletRepo.Put(sourceWallet, trx); err != nil {
			return err
		}

		if err := t.mutationRepo.Put(entity.Mutation{
			ID:       uuid.New().String(),
			WalletID: sourceWallet.ID,
			UserID:   user.ID,
			Type:     0, // down
			Amount:   amount,
		}, trx); err != nil {
			return err
		}

		if err := t.mutationRepo.Put(entity.Mutation{
			ID:       uuid.New().String(),
			WalletID: targetWallet.ID,
			UserID:   target.ID,
			Type:     1, // topup
			Amount:   amount,
		}, trx); err != nil {
			return err
		}

		return nil
	})
}
