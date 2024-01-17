package caching

import (
	"errors"
	"slices"
	"sync"

	"github.com/dgraph-io/ristretto"
)

// ! For setting please ALWAYS use cost 1
// Room ID -> Table
var tablesCache *ristretto.Cache

func setupTablesCache() {
	var err error
	tablesCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5,     // expecting to store 10k tables
		MaxCost:     1 << 30, // maximum cost of cache is 1GB
		BufferItems: 64,      // Some random number, check docs
		OnEvict: func(item *ristretto.Item) {
			// TODO: Implement
		},
	})

	if err != nil {
		panic(err)
	}
}

// Errors
var (
	ErrTableNotFound            = errors.New("table not found")
	ErrClientAlreadyJoinedTable = errors.New("client already joined table")
	ErrCouldntCreateTable       = errors.New("couldn't create table")
)

type TableData struct {
	Mutex   *sync.Mutex
	Room    string
	Members []string
	Objects *ristretto.Cache // Cache for all objects on the table (Object ID -> Object)
}

type TableObject struct {
	ID       string `json:"id"`
	Location string `json:"loc"` // x:y encoded or sth
	Type     string `json:"t"`
	Creator  string `json:"c"` // ID of the creator
	Holder   string `json:"h"` // ID of the current card holder (others can't move it while it's held)
	Data     string `json:"d"` // Encrypted
}

// * Table management
func JoinTable(room string, client string) error {

	obj, valid := tablesCache.Get(room)
	var table *TableData
	if !valid {

		// Create object cache
		objectCache, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1_000,      // expecting to store 1k objects
			MaxCost:     10_000_000, // maximum cost of cache is 10 MB
			BufferItems: 64,         // Some random number, check docs
			OnEvict: func(item *ristretto.Item) {
				// TODO: Implement
			},
		})
		if err != nil {
			return err
		}

		// Create table
		table = &TableData{
			Mutex:   &sync.Mutex{},
			Room:    room,
			Members: []string{},
			Objects: objectCache,
		}
		tablesCache.Set(room, table, 1)
	} else {
		table = obj.(*TableData)
	}

	table.Mutex.Lock()
	if slices.Contains(table.Members, client) {
		return ErrClientAlreadyJoinedTable
	}
	table.Members = append(table.Members, client)
	table.Mutex.Unlock()

	return nil
}

func GetTable(room string) (bool, *TableData) {
	obj, valid := tablesCache.Get(room)
	if !valid {
		return false, nil
	}
	return true, obj.(*TableData)
}

func TableMembers(room string) (bool, []string) {
	obj, valid := tablesCache.Get(room)
	if !valid {
		return false, nil
	}
	return true, obj.(*TableData).Members
}

func LeaveTable(room string, client string) error {
	obj, valid := tablesCache.Get(room)
	if !valid {
		return ErrTableNotFound
	}
	table := obj.(*TableData)

	table.Mutex.Lock()
	for i, member := range table.Members {
		if member == client {
			table.Members = append(table.Members[:i], table.Members[i+1:]...)
			break
		}
	}
	table.Mutex.Unlock()

	return nil
}

// * Object helpers
func GetObjectFromTable(room string, object TableObject) {
}

func AddObjectToTable(room string, object TableObject) {
}

func RemoveObjectFromTable(room string, object TableObject) {
}
