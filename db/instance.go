package db

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var ErrTableAlreadyExists = errors.New("table already exists")
var ErrTableIsNotFound = errors.New("table is not found")

var DefaultOperationLimit = 100

type Instance struct {
	tables                map[string]any
	operationChan         chan operationArgument
	operationWg           *sync.WaitGroup
	operationOpen         atomic.Bool
	transactionIdentifier string
}

type operationArgument struct {
	op        func(*Instance) error
	operation string
	result    chan error
}

func NewInstance() *Instance {
	return &Instance{
		tables:                map[string]any{},
		operationChan:         make(chan operationArgument, DefaultOperationLimit), // buffered allocation, faster since the memory is already allocated first instead of dynamically
		operationWg:           &sync.WaitGroup{},
		operationOpen:         atomic.Bool{},
		transactionIdentifier: "main",
	}
}

// Start database daemon
func (i *Instance) Start() {
	i.operationOpen.Store(true)

	for op := range i.operationChan {
		if !i.operationOpen.Load() {
			// operation already close in here
			continue
		}
		i.operationWg.Add(1)

		// Wrap it with function, to handle panic cases.
		err := func() (err2 error) {
			defer func() {
				if v := recover(); v != nil {
					err2 = fmt.Errorf("error %v", v)
				}
				i.operationWg.Done()
			}()
			return op.op(i)
		}()
		op.result <- err
	}
}

func (i *Instance) enqueueProcess(f func(*Instance) error, operationName string) error {
	opArgument := operationArgument{
		op:        f,
		result:    make(chan error, 1),
		operation: operationName,
	}
	i.operationChan <- opArgument

	return <-opArgument.result

}

func (i *Instance) Close() {
	// Close the booelan, so that wont be upcoming request
	i.operationOpen.Store(false)

	// Wait for in-progress request to finish
	i.operationWg.Wait()

	// Lastly, close the channel
	close(i.operationChan)
}

func (i *Instance) CreateTable(tableName string) error {
	// We use lambda function
	op := func(x *Instance) error {
		if _, ok := x.tables[tableName]; ok {
			return ErrTableAlreadyExists
		}

		// initialize table
		x.tables[tableName] = &sync.Map{}
		return nil
	}

	return i.enqueueProcess(op, "createTable")
}

func (i *Instance) GetTable(tableName string) (*Table, error) {
	table, found := i.tables[tableName]
	if !found {
		return nil, ErrTableIsNotFound
	}

	return &Table{
		data:           table.(*sync.Map),
		enqueueProcess: i.enqueueProcess,
	}, nil
}

func (i *Instance) Transaction(f func(*Transaction) error) error {
	op := func(x *Instance) error {
		transaction := &Transaction{
			tables:  x.tables,
			changes: make(map[string]map[any]any),
		}

		if err := f(transaction); err != nil {
			// rollback don't do anything
			return err
		}

		// commit
		for table, change := range transaction.changes {
			loadedTable := x.tables[table]
			assertedTable := loadedTable.(*sync.Map)
			for primaryKey, row := range change {
				assertedTable.Store(primaryKey, row)
			}
		}

		return nil
	}

	return i.enqueueProcess(op, "transaction")
}
