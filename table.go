package main

import (
	"fmt"
)

type DataType int

const (
	TypeInt DataType = iota
	TypeString
	TypeFloat
)

type Column struct {
	Name       string
	Type       DataType
	PrimaryKey bool
	Unique     bool
	NotNull    bool
}

type Row map[string]interface{}

type Table struct {
	Name       string
	Columns    []Column
	Rows       []Row
	nextID     int
	indexes    map[string]map[interface{}][]int // column -> value -> row indices
	primaryKey string
}

func NewTable(name string, columns []Column) *Table {
	t := &Table{
		Name:    name,
		Columns: columns,
		Rows:    make([]Row, 0),
		nextID:  1,
		indexes: make(map[string]map[interface{}][]int),
	}

	// Create indexes for primary and unique columns
	for _, col := range columns {
		if col.PrimaryKey {
			t.primaryKey = col.Name
			t.indexes[col.Name] = make(map[interface{}][]int)
		} else if col.Unique {
			t.indexes[col.Name] = make(map[interface{}][]int)
		}
	}

	return t
}

func (t *Table) Insert(values Row) error {
	// Validate columns
	row := make(Row)

	for _, col := range t.Columns {
		val, exists := values[col.Name]

		if !exists {
			if col.NotNull {
				return fmt.Errorf("column %s cannot be null", col.Name)
			}
			row[col.Name] = nil
			continue
		}

		// Type validation
		if err := t.validateType(col, val); err != nil {
			return err
		}

		// Check unique/primary key constraints
		if col.PrimaryKey || col.Unique {
			if t.valueExists(col.Name, val) {
				return fmt.Errorf("duplicate value for %s: %v", col.Name, val)
			}
		}

		row[col.Name] = val
	}

	// Add row
	rowIdx := len(t.Rows)
	t.Rows = append(t.Rows, row)

	// Update indexes
	for colName, index := range t.indexes {
		if val, exists := row[colName]; exists && val != nil {
			index[val] = append(index[val], rowIdx)
		}
	}

	return nil
}

func (t *Table) validateType(col Column, val interface{}) error {
	switch col.Type {
	case TypeInt:
		if _, ok := val.(int); !ok {
			return fmt.Errorf("invalid type for %s: expected int", col.Name)
		}
	case TypeString:
		if _, ok := val.(string); !ok {
			return fmt.Errorf("invalid type for %s: expected string", col.Name)
		}
	case TypeFloat:
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("invalid type for %s: expected float", col.Name)
		}
	}
	return nil
}

func (t *Table) valueExists(colName string, val interface{}) bool {
	if index, exists := t.indexes[colName]; exists {
		if _, found := index[val]; found {
			return true
		}
	}
	return false
}

func (t *Table) Select(columns []string, where *WhereClause) []Row {
	result := make([]Row, 0)

	// If WHERE clause uses indexed column with equality, use index
	if where != nil && where.Op == "=" {
		if index, indexed := t.indexes[where.Column]; indexed {
			if indices, found := index[where.Value]; found {
				for _, idx := range indices {
					if t.matchesWhere(t.Rows[idx], where) {
						result = append(result, t.projectRow(t.Rows[idx], columns))
					}
				}
				return result
			}
		}
	}

	// Full table scan
	for _, row := range t.Rows {
		if t.matchesWhere(row, where) {
			result = append(result, t.projectRow(row, columns))
		}
	}

	return result
}

func (t *Table) matchesWhere(row Row, where *WhereClause) bool {
	if where == nil {
		return true
	}

	val, exists := row[where.Column]
	if !exists {
		return false
	}

	switch where.Op {
	case "=":
		return val == where.Value
	case ">":
		return compareValues(val, where.Value) > 0
	case "<":
		return compareValues(val, where.Value) < 0
	case ">=":
		return compareValues(val, where.Value) >= 0
	case "<=":
		return compareValues(val, where.Value) <= 0
	case "!=":
		return val != where.Value
	}

	return false
}

func compareValues(a, b interface{}) int {
	switch av := a.(type) {
	case int:
		bv := b.(int)
		if av < bv {
			return -1
		} else if av > bv {
			return 1
		}
		return 0
	case float64:
		bv := b.(float64)
		if av < bv {
			return -1
		} else if av > bv {
			return 1
		}
		return 0
	case string:
		bv := b.(string)
		if av < bv {
			return -1
		} else if av > bv {
			return 1
		}
		return 0
	}
	return 0
}

func (t *Table) projectRow(row Row, columns []string) Row {
	if len(columns) == 1 && columns[0] == "*" {
		return row
	}

	result := make(Row)
	for _, col := range columns {
		if val, exists := row[col]; exists {
			result[col] = val
		}
	}
	return result
}

func (t *Table) Update(updates map[string]interface{}, where *WhereClause) int {
	count := 0

	for i, row := range t.Rows {
		if t.matchesWhere(row, where) {
			for col, val := range updates {
				// Remove old index entry
				if index, indexed := t.indexes[col]; indexed {
					if oldVal, exists := row[col]; exists {
						t.removeFromIndex(index, oldVal, i)
					}
				}

				row[col] = val

				// Add new index entry
				if index, indexed := t.indexes[col]; indexed {
					index[val] = append(index[val], i)
				}
			}
			count++
		}
	}

	return count
}

func (t *Table) Delete(where *WhereClause) int {
	newRows := make([]Row, 0)
	count := 0

	for _, row := range t.Rows {
		if t.matchesWhere(row, where) {
			count++
			// Remove from indexes
			for colName, index := range t.indexes {
				if val, exists := row[colName]; exists {
					t.removeFromIndex(index, val, len(newRows))
				}
			}
		} else {
			newRows = append(newRows, row)
		}
	}

	t.Rows = newRows

	// Rebuild indexes
	t.rebuildIndexes()

	return count
}

func (t *Table) removeFromIndex(index map[interface{}][]int, val interface{}, rowIdx int) {
	if indices, found := index[val]; found {
		newIndices := make([]int, 0)
		for _, idx := range indices {
			if idx != rowIdx {
				newIndices = append(newIndices, idx)
			}
		}
		if len(newIndices) > 0 {
			index[val] = newIndices
		} else {
			delete(index, val)
		}
	}
}

func (t *Table) rebuildIndexes() {
	// Clear indexes
	for colName := range t.indexes {
		t.indexes[colName] = make(map[interface{}][]int)
	}

	// Rebuild
	for i, row := range t.Rows {
		for colName, index := range t.indexes {
			if val, exists := row[colName]; exists && val != nil {
				index[val] = append(index[val], i)
			}
		}
	}
}

func (t *Table) Join(right *Table, leftCol, rightCol string, columns []string, where *WhereClause) []Row {
	result := make([]Row, 0)

	for _, leftRow := range t.Rows {
		leftVal := leftRow[leftCol]

		for _, rightRow := range right.Rows {
			if rightRow[rightCol] == leftVal {
				// Merge rows
				merged := make(Row)
				for k, v := range leftRow {
					merged[t.Name+"."+k] = v
				}
				for k, v := range rightRow {
					merged[right.Name+"."+k] = v
				}

				if t.matchesWhere(merged, where) {
					result = append(result, t.projectRow(merged, columns))
				}
			}
		}
	}

	return result
}
