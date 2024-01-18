package caching

import (
	"errors"
	"slices"
	"sync"

	"fajurion.com/voice-node/util"
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

// Configuration
const maxObjects = 100

// Errors
var (
	ErrTableNotFound            = errors.New("table not found")
	ErrClientAlreadyJoinedTable = errors.New("client already joined table")
	ErrCouldntCreateTable       = errors.New("couldn't create table")
	ErrObjectNotFound           = errors.New("one table object wasn't found")
)

type TableData struct {
	Mutex      *sync.Mutex
	Room       string
	Members    []string
	ObjectList []string         // List of all object ids
	Objects    *ristretto.Cache // Cache for all objects on the table (Object ID -> Object)
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

	if len(table.Members) == 0 {
		tablesCache.Del(room)
	}

	return nil
}

type TableObject struct {
	ID        string  `json:"id"`
	LocationX float64 `json:"x"`
	LocationY float64 `json:"y"`
	Type      string  `json:"t"`
	Creator   string  `json:"c"` // ID of the creator
	Holder    string  `json:"h"` // ID of the current card holder (others can't move it while it's held)
	Data      string  `json:"d"` // Encrypted
}

// * Object helpers
func AddObjectToTable(room string, object *TableObject) error {
	obj, valid := tablesCache.Get(room)
	if !valid {
		return ErrTableNotFound
	}
	table := obj.(*TableData)

	table.Mutex.Lock()

	// Generate random and unique id
	id := util.GenerateToken(5)
	for slices.Contains(table.ObjectList, id) {
		id = util.GenerateToken(5)
	}

	// Put object into cache and list
	table.ObjectList = append(table.ObjectList, id)
	object.ID = id
	table.Objects.Set(id, object, 1)
	table.Objects.Wait()

	table.Mutex.Unlock()

	return nil
}

func RemoveObjectFromTable(room string, object string) error {
	obj, valid := tablesCache.Get(room)
	if !valid {
		return ErrTableNotFound
	}
	table := obj.(*TableData)

	table.Mutex.Lock()

	// Put object into cache and list
	for i, member := range table.ObjectList {
		if member == object {
			table.ObjectList = append(table.ObjectList[:i], table.ObjectList[i+1:]...)
			break
		}
	}
	table.Objects.Del(object)
	table.Objects.Wait()

	table.Mutex.Unlock()

	return nil
}

func ModifyTableObject(room string, objectId string, data string) error {
	obj, valid := tablesCache.Get(room)
	if !valid {
		return ErrTableNotFound
	}
	table := obj.(*TableData)

	// Modify object data
	tObj, valid := table.Objects.Get(objectId)
	if !valid {
		return ErrObjectNotFound
	}
	object := tObj.(*TableObject)
	object.Data = data

	return nil
}

func MoveTableObject(room string, objectId string, x, y float64) error {
	obj, valid := tablesCache.Get(room)
	if !valid {
		return ErrTableNotFound
	}
	table := obj.(*TableData)

	// Modify object data
	tObj, valid := table.Objects.Get(objectId)
	if !valid {
		return ErrObjectNotFound
	}
	object := tObj.(*TableObject)
	object.LocationX = x
	object.LocationY = y

	return nil
}

func TableObjects(room string) ([]*TableObject, error) {
	obj, valid := tablesCache.Get(room)
	if !valid {
		return nil, ErrTableNotFound
	}
	table := obj.(*TableData)

	objects := make([]*TableObject, len(table.ObjectList))
	for i, value := range table.ObjectList {
		object, valid := table.Objects.Get(value)
		if !valid {
			return nil, ErrObjectNotFound
		}

		objects[i] = object.(*TableObject)
	}

	return objects, nil
}
