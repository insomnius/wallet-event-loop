package db

import "sync"

type Transaction struct {
	tables  *sync.Map
	changes map[string]map[any]any
}

func (t *Transaction) GetTable(tableName string) (*Table, error) {
	table, found := t.tables.Load(tableName)
	if !found {
		return nil, ErrTableIsNotFound
	}

	if _, found := t.changes[tableName]; !found {
		t.changes[tableName] = make(map[any]any)
	}

	// identification use only
	clonedInstance := NewInstance()
	clonedInstance.transactionIdentifier = "sub"

	return &Table{
		data: table.(*sync.Map),
		enqueueProcess: func(f func(*Instance) error, operationName string) error {
			return f(clonedInstance)
		},
		changes: t.changes[tableName],
	}, nil
}
