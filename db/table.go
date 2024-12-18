package db

import (
	"errors"
)

var ErrNotFound = errors.New("not found")

type Table struct {
	data           map[string]any
	changes        map[string]any
	enqueueProcess func(f func(*Instance) error, operationName string) error
}

func (t *Table) FindByID(id string) (any, error) {
	var v any
	var found bool
	if changeV, ok := t.changes[id]; ok {
		v = changeV
		found = true
		return v, nil
	}

	// read uncommitted
	v, found = t.data[id]
	if !found {
		return nil, ErrNotFound
	}

	return v, nil
}

func (t *Table) Filter(f func(v any) bool) []any {
	filtered := []any{}

	for key, value := range t.data {
		// handling read commited
		if changeV, ok := t.changes[key]; ok {
			value = changeV
		}

		// read uncommitted
		if f(value) {
			filtered = append(filtered, value)
		}
	}

	return filtered
}

func (t *Table) ReplaceOrStore(id string, value any) any {
	var v any
	op := func(i *Instance) error {

		// handling write uncommited
		if i.transactionIdentifier == "sub" {
			t.changes[id] = value
			return nil
		}

		// handling write commited
		t.data[id] = value
		return nil
	}

	_ = t.enqueueProcess(op, "replaceOrStore")
	return v
}
