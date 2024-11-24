package db

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Table struct {
	data           *sync.Map
	changes        map[any]any
	enqueueProcess func(f func(*Instance) error, operationName string) error
}

func (t *Table) FindByID(id string) (any, error) {
	var v any
	var found bool
	err := t.enqueueProcess(func(*Instance) error {
		if changeV, ok := t.changes[id]; ok {
			v = changeV
			found = true
			return nil
		}

		v, found = t.data.Load(id)
		if !found {
			return ErrNotFound
		}

		return nil
	}, "findById")
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (t *Table) Filter(f func(v any) bool) []any {
	filtered := []any{}

	_ = t.enqueueProcess(func(*Instance) error {
		t.data.Range(func(key, value any) bool {
			if changeV, ok := t.changes[key]; ok {
				value = changeV
			}

			if f(value) {
				filtered = append(filtered, value)
			}

			return true
		})

		return nil
	}, "filter")

	return filtered
}

func (t *Table) ReplaceOrStore(id string, value any) any {
	var v any
	op := func(i *Instance) error {

		if i.transactionIdentifier == "sub" {
			t.changes[id] = value
			return nil
		}

		v, _ = t.data.LoadOrStore(id, value)
		return nil
	}

	_ = t.enqueueProcess(op, "replaceOrStore")
	return v
}
