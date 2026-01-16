package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type PersistenceManager struct {
	filepath string
	mu       sync.RWMutex
}

func NewPersistenceManager(filepath string) *PersistenceManager {
	return &PersistenceManager{filepath: filepath}
}

func (pm *PersistenceManager) Save(db *Database) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data := make(map[string]interface{})
	for name, table := range db.tables {
		data[name] = map[string]interface{}{
			"columns": table.Columns,
			"rows":    table.Rows,
		}
	}

	file, err := os.Create(pm.filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (pm *PersistenceManager) Load(db *Database) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	file, err := os.Open(pm.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode data: %v", err)
	}

	for name, tableData := range data {
		td := tableData.(map[string]interface{})

		var columns []Column
		colData, _ := json.Marshal(td["columns"])
		err := json.Unmarshal(colData, &columns)
		if err != nil {
			return err
		}

		table := NewTable(name, columns)

		var rows []Row
		rowData, _ := json.Marshal(td["rows"])
		err = json.Unmarshal(rowData, &rows)
		if err != nil {
			return err
		}
		table.Rows = rows

		table.rebuildIndexes()
		db.tables[name] = table
	}

	return nil
}
