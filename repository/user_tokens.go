package repository

import (
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
)

type UserToken struct {
	db *db.Instance
}

func NewUserToken(db *db.Instance) *UserToken {
	return &UserToken{
		db: db,
	}
}

func (u *UserToken) FindByToken(token string, txs ...*db.Transaction) (entity.UserToken, error) {
	t, err := u.table(txs...)
	if err != nil {
		return entity.UserToken{}, err
	}

	v, err := t.FindByID(token)
	if err != nil {
		return entity.UserToken{}, err
	}

	return v.(entity.UserToken), nil
}

func (u *UserToken) Put(userToken entity.UserToken, txs ...*db.Transaction) error {
	t, err := u.table(txs...)
	if err != nil {
		return err
	}

	_ = t.ReplaceOrStore(userToken.Token, userToken)

	return nil
}

func (u *UserToken) table(txs ...*db.Transaction) (*db.Table, error) {
	// if there is open transactions
	// then the db will use transactions instead of default db connections
	if len(txs) > 0 {
		return txs[0].GetTable("user_tokens")
	}
	return u.db.GetTable("user_tokens")
}
