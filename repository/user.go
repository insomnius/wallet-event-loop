package repository

import (
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
)

type User struct {
	db *db.Instance
}

func NewUser(db *db.Instance) *User {
	return &User{
		db: db,
	}
}

func (u *User) FindById(id string, txs ...*db.Transaction) (*entity.User, error) {
	t, err := u.table(txs...)
	if err != nil {
		return nil, err
	}

	v, err := t.FindByID(id)
	if err != nil {
		return nil, err
	}

	return v.(*entity.User), nil
}

func (u *User) FindByEmail(email string, txs ...*db.Transaction) (*entity.User, error) {
	t, err := u.table(txs...)
	if err != nil {
		return nil, err
	}

	v := t.Filter(func(v any) bool {
		return v.(*entity.User).Email == email
	})
	if len(v) == 0 {
		return nil, db.ErrNotFound
	}

	return v[0].(*entity.User), nil
}

func (u *User) Put(user *entity.User, txs ...*db.Transaction) error {
	t, err := u.table(txs...)
	if err != nil {
		return err
	}

	_ = t.ReplaceOrStore(user.ID, user)

	return nil
}

func (u *User) table(txs ...*db.Transaction) (*db.Table, error) {
	// if there is open transactions
	// then the db will use transactions instead of default db connections
	if len(txs) > 0 {
		return txs[0].GetTable("users")
	}
	return u.db.GetTable("users")
}
