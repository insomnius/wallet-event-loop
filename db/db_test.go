package db_test

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
)

func TestCreateMultiple(t *testing.T) {
	inst := db.NewInstance()
	defer inst.Close()

	go func() {
		inst.Start()
	}()

	inst.CreateTable("user")
	table, _ := inst.GetTable("user")

	table.ReplaceOrStore("xx", entity.User{
		ID:   "xx",
		Name: "super",
	})
	table.ReplaceOrStore("yy", entity.User{
		ID:   "yy",
		Name: "super",
	})
}

func TestTransaction(t *testing.T) {
	inst := db.NewInstance()
	defer inst.Close()

	go func() {
		inst.Start()
	}()

	inst.CreateTable("users")

	var errCount int32
	var successCount int32

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := inst.Transaction(func(x *db.Instance) error {
				userTable, err := x.GetTable("users")
				if err != nil {
					return err
				}

				_, err = userTable.FindByID("xx")
				if err == nil {
					return errors.New("data dengan id xx sudah ada")
				}

				userTable.ReplaceOrStore("xx", entity.User{
					ID:   "xx",
					Name: "super",
				})

				return nil
			})
			if err != nil {
				atomic.AddInt32(&errCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()

	if errCount != 9 {
		t.Fatal("transaction process failed", errCount, successCount)
	}

	if successCount != 1 {
		t.Fatal("transaction process failed", errCount, successCount)
	}
}

func BenchmarkCreateMultiple(b *testing.B) {
	inst := db.NewInstance()
	defer inst.Close()

	go func() {
		inst.Start()
	}()

	inst.CreateTable("user")
	table, _ := inst.GetTable("user")

	for i := 0; i < b.N; i++ {
		table.ReplaceOrStore(strconv.Itoa(i), entity.User{
			ID:   strconv.Itoa(i),
			Name: strconv.Itoa(i),
		})
	}
}
