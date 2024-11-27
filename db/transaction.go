package db

type Transaction struct {
	tables  map[string]map[string]any
	changes map[string]map[string]any
}

func (t *Transaction) GetTable(tableName string) (*Table, error) {
	table, found := t.tables[tableName]
	if !found {
		return nil, ErrTableIsNotFound
	}

	if _, found := t.changes[tableName]; !found {
		t.changes[tableName] = make(map[string]any)
	}

	// identification use only
	clonedInstance := NewInstance()
	clonedInstance.transactionIdentifier = "sub"

	return &Table{
		data: table,
		enqueueProcess: func(f func(*Instance) error, operationName string) error {
			return f(clonedInstance)
		},
		changes: t.changes[tableName],
	}, nil
}
