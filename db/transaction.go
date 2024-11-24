package db

import "sync"

type Transaction struct {
	tables *sync.Map
}

func (t *Transaction) GetTable(tableName string) (*Table, error) {
	table, found := t.tables.Load(tableName)
	if !found {
		return nil, ErrTableIsNotFound
	}

	return &Table{
		data: table.(*sync.Map),
		enqueueProcess: func(f func(*Instance) error, operationName string) error {
			clonedInstance := NewInstance()
			clonedInstance.tables = t.tables
			return f(clonedInstance)
		},
	}, nil
}