package main

import (
	"fmt"
	"strings"
)

type Database struct {
	tables map[string]*Table
}

type QueryResult struct {
	Columns []string
	Rows    []Row
	Message string
}

func NewDatabase() *Database {
	return &Database{
		tables: make(map[string]*Table),
	}
}

func (db *Database) Execute(query string) (*QueryResult, error) {
	stmt, err := Parse(query)
	if err != nil {
		return nil, err
	}

	switch s := stmt.(type) {
	case *CreateTableStmt:
		return db.executeCreate(s)
	case *InsertStmt:
		return db.executeInsert(s)
	case *SelectStmt:
		return db.executeSelect(s)
	case *UpdateStmt:
		return db.executeUpdate(s)
	case *DeleteStmt:
		return db.executeDelete(s)
	default:
		return nil, fmt.Errorf("unknown statement type")
	}
}

func (db *Database) executeCreate(stmt *CreateTableStmt) (*QueryResult, error) {
	if _, exists := db.tables[stmt.Name]; exists {
		return nil, fmt.Errorf("table %s already exists", stmt.Name)
	}

	db.tables[stmt.Name] = NewTable(stmt.Name, stmt.Columns)
	return &QueryResult{Message: fmt.Sprintf("Table %s created", stmt.Name)}, nil
}

func (db *Database) executeInsert(stmt *InsertStmt) (*QueryResult, error) {
	table, exists := db.tables[stmt.Table]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", stmt.Table)
	}

	if err := table.Insert(stmt.Values); err != nil {
		return nil, err
	}

	return &QueryResult{Message: "1 row inserted"}, nil
}

func (db *Database) executeSelect(stmt *SelectStmt) (*QueryResult, error) {
	table, exists := db.tables[stmt.Table]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", stmt.Table)
	}

	rows := table.Select(stmt.Columns, stmt.Where)

	columns := stmt.Columns
	if len(columns) == 1 && columns[0] == "*" {
		columns = make([]string, len(table.Columns))
		for i, col := range table.Columns {
			columns[i] = col.Name
		}
	}

	return &QueryResult{Columns: columns, Rows: rows}, nil
}

func (db *Database) executeUpdate(stmt *UpdateStmt) (*QueryResult, error) {
	table, exists := db.tables[stmt.Table]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", stmt.Table)
	}

	count := table.Update(stmt.Updates, stmt.Where)
	return &QueryResult{Message: fmt.Sprintf("%d row(s) updated", count)}, nil
}

func (db *Database) executeDelete(stmt *DeleteStmt) (*QueryResult, error) {
	table, exists := db.tables[stmt.Table]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", stmt.Table)
	}

	count := table.Delete(stmt.Where)
	return &QueryResult{Message: fmt.Sprintf("%d row(s) deleted", count)}, nil
}

func (r *QueryResult) Print() {
	if r.Message != "" {
		fmt.Println(r.Message)
		return
	}

	if len(r.Rows) == 0 {
		fmt.Println("No results")
		return
	}

	// Print header
	for i, col := range r.Columns {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Printf("%-15s", col)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", len(r.Columns)*18))

	// Print rows
	for _, row := range r.Rows {
		for i, col := range r.Columns {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Printf("%-15v", row[col])
		}
		fmt.Println()
	}
	fmt.Printf("\n%d row(s)\n", len(r.Rows))
}
