package db

import (
	"errors"
	"sync"
)

var ErrTableAlreadyExists = errors.New("table already exists")
var ErrTableIsNotFound = errors.New("table is not found")

var DefaultOperationLimit = 100

type Instance struct {
	tables        map[string]*sync.Map
	operationChan chan operationArgument
	operationWg   *sync.WaitGroup
	operationOpen bool
}

type operationArgument struct {
	op     func() error
	result chan error
}

func NewInstance() *Instance {
	return &Instance{
		tables:        map[string]*sync.Map{},
		operationChan: make(chan operationArgument, DefaultOperationLimit), // buffered allocation, faster since the memory is already allocated first instead of dynamically
		operationWg:   &sync.WaitGroup{},
		operationOpen: false,
	}
}

// Start database daemon
func (i *Instance) Start() {
	i.operationOpen = true

	for op := range i.operationChan {
		if !i.operationOpen {
			// operation already close in here
			continue
		}
		i.operationWg.Add(1)

		// Wrap it with function, to handle panic cases.
		err := func() error {
			defer func() {
				recover()
				i.operationWg.Done()
			}()

			return op.op()
		}()
		op.result <- err
	}
}

func (i *Instance) enqueueProcess(f func() error) error {
	opArgument := operationArgument{
		op:     f,
		result: make(chan error, 1),
	}
	i.operationChan <- opArgument

	return <-opArgument.result

}

func (i *Instance) Close() {
	// Close the booelan, so that wont be upcoming request
	i.operationOpen = false

	// Wait for in-progress request to finish
	i.operationWg.Wait()

	// Lastly, close the channel
	close(i.operationChan)
}

func (i *Instance) CreateTable(tableName string) error {
	// We use lambda function
	op := func() error {
		if _, ok := i.tables[tableName]; ok {
			return ErrTableAlreadyExists
		}

		// initialize table
		i.tables[tableName] = &sync.Map{}
		return nil
	}

	return i.enqueueProcess(op)
}

func (i *Instance) GetTable(tableName string) (*Table, error) {
	table, found := i.tables[tableName]
	if !found {
		return nil, ErrTableIsNotFound
	}

	return &Table{
		data: table,
	}, nil
}

func (i *Instance) Transaction() {

}
