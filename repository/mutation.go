package repository

import (
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
)

type Mutation struct {
	db *db.Instance
}

func NewMutation(db *db.Instance) *Mutation {
	return &Mutation{
		db: db,
	}
}

func (u *Mutation) FindById(id string, txs ...*db.Transaction) (*entity.Mutation, error) {
	t, err := u.table(txs...)
	if err != nil {
		return nil, err
	}

	v, err := t.FindByID(id)
	if err != nil {
		return nil, err
	}

	return v.(*entity.Mutation), nil
}

func (u *Mutation) Create(mutation *entity.Mutation, txs ...*db.Transaction) error {
	t, err := u.table(txs...)
	if err != nil {
		return err
	}

	_ = t.ReplaceOrStore(mutation.ID, mutation)

	return nil
}

func (u *Mutation) table(txs ...*db.Transaction) (*db.Table, error) {
	// if there is open transactions
	// then the db will use transactions instead of default db connections
	if len(txs) > 0 {
		return txs[0].GetTable("mutations")
	}
	return u.db.GetTable("mutations")
}
