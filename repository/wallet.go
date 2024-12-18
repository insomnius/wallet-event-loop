package repository

import (
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
)

type Wallet struct {
	db *db.Instance
}

func NewWallet(db *db.Instance) *Wallet {
	return &Wallet{
		db: db,
	}
}

func (u *Wallet) FindById(id string, txs ...*db.Transaction) (entity.Wallet, error) {
	t, err := u.table(txs...)
	if err != nil {
		return entity.Wallet{}, err
	}

	v, err := t.FindByID(id)
	if err != nil {
		return entity.Wallet{}, err
	}

	return v.(entity.Wallet), nil
}

func (u *Wallet) FindByUserID(userID string, txs ...*db.Transaction) (entity.Wallet, error) {
	t, err := u.table(txs...)
	if err != nil {
		return entity.Wallet{}, err
	}

	filtered := t.Filter(func(v any) bool {
		return v.(entity.Wallet).UserID == userID
	})

	if len(filtered) == 0 {
		return entity.Wallet{}, db.ErrNotFound
	}

	return filtered[0].(entity.Wallet), nil
}

func (u *Wallet) Put(wallet entity.Wallet, txs ...*db.Transaction) error {
	t, err := u.table(txs...)
	if err != nil {
		return err
	}

	_ = t.ReplaceOrStore(wallet.ID, wallet)

	return nil
}

func (u *Wallet) table(txs ...*db.Transaction) (*db.Table, error) {
	// if there is open transactions
	// then the db will use transactions instead of default db connections
	if len(txs) > 0 {
		return txs[0].GetTable("wallets")
	}
	return u.db.GetTable("wallets")
}
