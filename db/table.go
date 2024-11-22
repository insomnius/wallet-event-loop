package db

import "sync"

type Table struct {
	data *sync.Map
}
