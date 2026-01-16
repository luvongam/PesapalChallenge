package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Statement interface{}

type CreateTableStmt struct {
	Name    string
	Columns []Column
}

type InsertStmt struct {
	Table  string
	Values Row
}

type SelectStmt struct {
	Columns []string
	Table   string
	Where   *WhereClause
	Join    *JoinClause
}

type UpdateStmt struct {
	Table   string
	Updates map[string]interface{}
	Where   *WhereClause
}

type DeleteStmt struct {
	Table string
	Where *WhereClause
}

type WhereClause struct {
	Column string
	Op     string
	Value  interface{}
}

type JoinClause struct {
	Table    string
	LeftCol  string
	RightCol string
}

func Parse(query string) (Statement, error) {
	query = strings.TrimSpace(query)
	tokens := tokenize(query)

	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty query")
	}

	switch strings.ToUpper(tokens[0]) {
	case "CREATE":
		return parseCreateTable(tokens)
	case "INSERT":
		return parseInsert(tokens)
	case "SELECT":
		return parseSelect(tokens)
	case "UPDATE":
		return parseUpdate(tokens)
	case "DELETE":
		return parseDelete(tokens)
	default:
		return nil, fmt.Errorf("unsupported command: %s", tokens[0])
	}
}

func tokenize(query string) []string {
	// Simple tokenizer - splits on spaces but preserves quoted strings
	tokens := make([]string, 0)
	current := ""
	inQuotes := false

	for i := 0; i < len(query); i++ {
		ch := query[i]

		if ch == '\'' || ch == '"' {
			inQuotes = !inQuotes
		} else if (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r') && !inQuotes {
			if current != "" {
				tokens = append(tokens, current)
				current = ""
			}
		} else if (ch == '(' || ch == ')' || ch == ',' || ch == ';') && !inQuotes {
			if current != "" {
				tokens = append(tokens, current)
				current = ""
			}
			if ch != ',' && ch != ';' {
				tokens = append(tokens, string(ch))
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		tokens = append(tokens, current)
	}

	return tokens
}

func parseCreateTable(tokens []string) (*CreateTableStmt, error) {
	// CREATE TABLE tablename (col1 TYPE constraints, col2 TYPE constraints)
	if len(tokens) < 5 || strings.ToUpper(tokens[1]) != "TABLE" {
		return nil, fmt.Errorf("invalid CREATE TABLE syntax")
	}

	stmt := &CreateTableStmt{
		Name:    tokens[2],
		Columns: make([]Column, 0),
	}

	// Find opening paren
	parenIdx := -1
	for i, tok := range tokens {
		if tok == "(" {
			parenIdx = i
			break
		}
	}

	if parenIdx == -1 {
		return nil, fmt.Errorf("missing column definitions")
	}

	// Parse columns
	i := parenIdx + 1
	for i < len(tokens) && tokens[i] != ")" {
		col := Column{
			Name: tokens[i],
		}
		i++

		if i >= len(tokens) {
			return nil, fmt.Errorf("incomplete column definition")
		}

		// Parse type
		typeStr := strings.ToUpper(tokens[i])
		switch typeStr {
		case "INT", "INTEGER":
			col.Type = TypeInt
		case "STRING", "VARCHAR", "TEXT":
			col.Type = TypeString
		case "FLOAT", "REAL":
			col.Type = TypeFloat
		default:
			return nil, fmt.Errorf("unknown type: %s", typeStr)
		}
		i++

		// Parse constraints
		for i < len(tokens) && tokens[i] != ")" {
			constraint := strings.ToUpper(tokens[i])
			if constraint == "PRIMARY" && i+1 < len(tokens) && strings.ToUpper(tokens[i+1]) == "KEY" {
				col.PrimaryKey = true
				i += 2
			} else if constraint == "UNIQUE" {
				col.Unique = true
				i++
			} else if constraint == "NOT" && i+1 < len(tokens) && strings.ToUpper(tokens[i+1]) == "NULL" {
				col.NotNull = true
				i += 2
			} else {
				break
			}
		}

		stmt.Columns = append(stmt.Columns, col)
	}

	return stmt, nil
}

func parseInsert(tokens []string) (*InsertStmt, error) {
	// INSERT INTO tablename (col1, col2) VALUES (val1, val2)
	if len(tokens) < 4 || strings.ToUpper(tokens[1]) != "INTO" {
		return nil, fmt.Errorf("invalid INSERT syntax")
	}

	stmt := &InsertStmt{
		Table:  tokens[2],
		Values: make(Row),
	}

	// Find column names
	cols := make([]string, 0)
	i := 3
	if i < len(tokens) && tokens[i] == "(" {
		i++
		for i < len(tokens) && tokens[i] != ")" {
			cols = append(cols, tokens[i])
			i++
		}
		i++ // skip )
	}

	// Find VALUES
	for i < len(tokens) && strings.ToUpper(tokens[i]) != "VALUES" {
		i++
	}
	i++ // skip VALUES

	// Parse values
	if i >= len(tokens) || tokens[i] != "(" {
		return nil, fmt.Errorf("missing VALUES")
	}
	i++ // skip (

	vals := make([]interface{}, 0)
	for i < len(tokens) && tokens[i] != ")" {
		val := parseValue(tokens[i])
		vals = append(vals, val)
		i++
	}

	// Match columns to values
	for j, col := range cols {
		if j < len(vals) {
			stmt.Values[col] = vals[j]
		}
	}

	return stmt, nil
}

func parseSelect(tokens []string) (*SelectStmt, error) {
	// SELECT col1, col2 FROM table WHERE col = val
	// SELECT * FROM table1 JOIN table2 ON table1.id = table2.id
	stmt := &SelectStmt{
		Columns: make([]string, 0),
	}

	i := 1
	// Parse columns
	for i < len(tokens) && strings.ToUpper(tokens[i]) != "FROM" {
		stmt.Columns = append(stmt.Columns, tokens[i])
		i++
	}

	if i >= len(tokens) || strings.ToUpper(tokens[i]) != "FROM" {
		return nil, fmt.Errorf("missing FROM clause")
	}
	i++ // skip FROM

	if i >= len(tokens) {
		return nil, fmt.Errorf("missing table name")
	}
	stmt.Table = tokens[i]
	i++

	// Check for JOIN
	if i < len(tokens) && strings.ToUpper(tokens[i]) == "JOIN" {
		i++
		if i >= len(tokens) {
			return nil, fmt.Errorf("missing join table")
		}

		join := &JoinClause{
			Table: tokens[i],
		}
		i++

		if i < len(tokens) && strings.ToUpper(tokens[i]) == "ON" {
			i++
			// Parse ON condition: table1.col = table2.col
			if i+2 < len(tokens) {
				leftParts := strings.Split(tokens[i], ".")
				if len(leftParts) == 2 {
					join.LeftCol = leftParts[0] + "." + leftParts[1]
				}
				i++ // skip =
				i++
				rightParts := strings.Split(tokens[i], ".")
				if len(rightParts) == 2 {
					join.RightCol = rightParts[0] + "." + rightParts[1]
				}
				i++
			}
		}

		stmt.Join = join
	}

	// Parse WHERE
	if i < len(tokens) && strings.ToUpper(tokens[i]) == "WHERE" {
		stmt.Where = parseWhere(tokens[i+1:])
	}

	return stmt, nil
}

func parseUpdate(tokens []string) (*UpdateStmt, error) {
	// UPDATE table SET col1 = val1, col2 = val2 WHERE col = val
	if len(tokens) < 4 {
		return nil, fmt.Errorf("invalid UPDATE syntax")
	}

	stmt := &UpdateStmt{
		Table:   tokens[1],
		Updates: make(map[string]interface{}),
	}

	// Find SET
	i := 2
	if strings.ToUpper(tokens[i]) != "SET" {
		return nil, fmt.Errorf("missing SET clause")
	}
	i++

	// Parse updates
	for i < len(tokens) && strings.ToUpper(tokens[i]) != "WHERE" {
		col := tokens[i]
		i++ // skip =
		i++
		val := parseValue(tokens[i])
		stmt.Updates[col] = val
		i++
	}

	// Parse WHERE
	if i < len(tokens) && strings.ToUpper(tokens[i]) == "WHERE" {
		stmt.Where = parseWhere(tokens[i+1:])
	}

	return stmt, nil
}

func parseDelete(tokens []string) (*DeleteStmt, error) {
	// DELETE FROM table WHERE col = val
	if len(tokens) < 3 || strings.ToUpper(tokens[1]) != "FROM" {
		return nil, fmt.Errorf("invalid DELETE syntax")
	}

	stmt := &DeleteStmt{
		Table: tokens[2],
	}

	// Parse WHERE
	if len(tokens) > 3 && strings.ToUpper(tokens[3]) == "WHERE" {
		stmt.Where = parseWhere(tokens[4:])
	}

	return stmt, nil
}

func parseWhere(tokens []string) *WhereClause {
	if len(tokens) < 3 {
		return nil
	}

	return &WhereClause{
		Column: tokens[0],
		Op:     tokens[1],
		Value:  parseValue(tokens[2]),
	}
}

func parseValue(s string) interface{} {
	// Try int
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}

	// Try float
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}

	// String (remove quotes if present)
	s = strings.Trim(s, "'\"")
	return s
}
