package db

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var ErrTableAlreadyExists = errors.New("table already exists")
var ErrTableIsNotFound = errors.New("table is not found")
var ErrInstanceAlreadyClosed = errors.New("database instance already closed")
var DefaultOperationLimit = 100

type Instance struct {
	tables                map[string]map[string]any
	operationWg           *sync.WaitGroup
	operationStop         atomic.Bool
	transactionIdentifier string
	operationStack        atomic.Pointer[[]*operationArgument]
}

type operationArgument struct {
	op        func(*Instance) error
	operation string
	wg        *sync.WaitGroup
	err       error
}

func NewInstance() *Instance {
	inst := &Instance{
		tables:                map[string]map[string]any{},
		operationWg:           &sync.WaitGroup{},
		operationStop:         atomic.Bool{},
		operationStack:        atomic.Pointer[[]*operationArgument]{},
		transactionIdentifier: "main",
	}
	go func() {
		inst.start()
	}()

	inst.operationStack.Store(&[]*operationArgument{})
	return inst
}

// Start database daemon
func (i *Instance) start() {
	// maxWait := time.Millisecond
	// waitTime := time.Duration(math.Min(float64(100*time.Nanosecond)*2, float64(maxWait)))

	for {
		if i.operationStop.Load() {
			break
		}

		currentOp := i.operationStack.Load()

		// waiting for new operations
		if len((*currentOp)) == 0 {
			// waitTime = time.Duration(min(float64(waitTime)*2, float64(maxWait)))
			time.Sleep(time.Nanosecond * 2)
			continue
		}
		// fmt.Println("(*currentOp)", (*currentOp), len((*currentOp)))
		// fmt.Println("KERJA")

		i.operationWg.Add(1)

		op := (*currentOp)[0]

		var remainingOp []*operationArgument
		if len((*currentOp)) > 0 {
			remainingOp = (*currentOp)[1:]
		} else {
			remainingOp = []*operationArgument{}
		}

		// Compare and swap until it success
		for !i.operationStack.CompareAndSwap(currentOp, &remainingOp) {
			currentOp = i.operationStack.Load()
			op = (*currentOp)[0]
			if len((*currentOp)) > 0 {
				remainingOp = (*currentOp)[1:]
			} else {
				remainingOp = []*operationArgument{}
			}
		}

		err := func() (errOp error) {
			defer func() {
				if v := recover(); v != nil {
					errOp = fmt.Errorf("error %v", v)
				}
				i.operationWg.Done()
			}()
			return op.op(i)
		}()
		op.err = err
		op.wg.Done()
	}
}

func (i *Instance) enqueueProcess(f func(*Instance) error, operationName string) error {
	if i.operationStop.Load() {
		return ErrInstanceAlreadyClosed
	}

	opArgument := &operationArgument{
		op:        f,
		operation: operationName,
		wg:        &sync.WaitGroup{},
	}
	opArgument.wg.Add(1)

	currentOp := i.operationStack.Load()
	newOpStack := append((*currentOp), opArgument)

	// Compare and swap until it success
	for !i.operationStack.CompareAndSwap(currentOp, &newOpStack) {
		currentOp = i.operationStack.Load()
		newOpStack = append((*currentOp), opArgument)
	}

	// fmt.Println("menunggu hasil", (*currentOp))

	opArgument.wg.Wait()
	return opArgument.err
}

func (i *Instance) Close() {
	// Close the booelan, so that wont be upcoming request
	i.operationStop.Store(true)

	// Wait for in-progress request to finish
	i.operationWg.Wait()
}

func (i *Instance) CreateTable(tableName string) error {
	// We use lambda function
	op := func(x *Instance) error {
		if _, ok := x.tables[tableName]; ok {
			return ErrTableAlreadyExists
		}

		// initialize table
		// x.tables[tableName] = &sync.Map{}
		x.tables[tableName] = map[string]any{}
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
		data:           table,
		enqueueProcess: i.enqueueProcess,
	}, nil
}

func (i *Instance) Transaction(f func(*Transaction) error) error {
	op := func(x *Instance) error {
		transaction := &Transaction{
			tables:  x.tables,
			changes: make(map[string]map[string]any),
		}

		if err := f(transaction); err != nil {
			// rollback don't do anything
			return err
		}

		// commit
		for table, change := range transaction.changes {
			assertedTable := x.tables[table]

			for primaryKey, row := range change {
				assertedTable[primaryKey] = row
			}
		}

		return nil
	}

	return i.enqueueProcess(op, "transaction")
}
