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

func (u *Mutation) FindById(id string, txs ...*db.Transaction) (entity.Mutation, error) {
	t, err := u.table(txs...)
	if err != nil {
		return entity.Mutation{}, err
	}

	v, err := t.FindByID(id)
	if err != nil {
		return entity.Mutation{}, err
	}

	return v.(entity.Mutation), nil
}

func (u *Mutation) Put(mutation entity.Mutation, txs ...*db.Transaction) error {
	t, err := u.table(txs...)
	if err != nil {
		return err
	}

	_ = t.ReplaceOrStore(mutation.ID, mutation)

	return nil
}

func (u *Mutation) GetByUserID(userID string, txs ...*db.Transaction) ([]entity.Mutation, error) {
	t, err := u.table(txs...)
	if err != nil {
		return nil, err
	}

	filtered := t.Filter(func(v any) bool {
		return v.(entity.Mutation).UserID == userID
	})

	if len(filtered) == 0 {
		return nil, db.ErrNotFound
	}

	converted := []entity.Mutation{}

	for _, v := range filtered {
		converted = append(converted, v.(entity.Mutation))
	}

	return converted, nil
}

func (u *Mutation) table(txs ...*db.Transaction) (*db.Table, error) {
	// if there is open transactions
	// then the db will use transactions instead of default db connections
	if len(txs) > 0 {
		return txs[0].GetTable("mutations")
	}
	return u.db.GetTable("mutations")
}
