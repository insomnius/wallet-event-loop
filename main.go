package main

import (
	"fmt"

	"github.com/insomnius/wallet-event-loop/db"
)

func main() {
	fmt.Println("Starting e-wallet services...")
	dbInstance := db.NewInstance()

	// Starting database instance
	go func() {
		dbInstance.Start()
	}()

	dbInstance.Close()

	fmt.Println("Closing e-wallet services...")
}
