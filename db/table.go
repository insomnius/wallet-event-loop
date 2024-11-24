package db

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Table struct {
	data           *sync.Map
	enqueueProcess func(f func(*Instance) error, operationName string) error
}

func (t *Table) FindByID(id string) (any, error) {
	v, found := t.data.Load(id)
	if !found {
		return nil, ErrNotFound
	}

	return v, nil
}

func (t *Table) ReplaceOrStore(id string, value any) any {
	var v any
	op := func(*Instance) error {
		v, _ = t.data.LoadOrStore(id, value)
		return nil
	}

	_ = t.enqueueProcess(op, "replaceOrStore")
	return v
}
