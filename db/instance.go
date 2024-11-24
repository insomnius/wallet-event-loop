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
	tables                *sync.Map
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
		tables:                &sync.Map{},
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
		if _, ok := x.tables.Load(tableName); ok {
			return ErrTableAlreadyExists
		}

		// initialize table
		x.tables.Store(tableName, &sync.Map{})
		return nil
	}

	return i.enqueueProcess(op, "createTable")
}

func (i *Instance) GetTable(tableName string) (*Table, error) {
	table, found := i.tables.Load(tableName)
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
		tables := &sync.Map{}

		x.tables.Range(func(key, value any) bool {
			tables.Store(key, cloneSyncMap(value.(*sync.Map)))
			return true
		})

		transaction := &Transaction{
			tables: tables,
		}

		if err := f(transaction); err != nil {
			return err
		}
		x.tables = transaction.tables
		return nil
	}

	return i.enqueueProcess(op, "transaction")
}

func cloneSyncMap(src *sync.Map) *sync.Map {
	dest := &sync.Map{}
	src.Range(func(key, value any) bool {
		dest.Store(key, value) // Copy each key-value pair
		return true
	})
	return dest
}
