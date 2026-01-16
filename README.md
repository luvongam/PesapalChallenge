# SimpleDB RDBMS

A lightweight relational database management system built from scratch in Go.

## Features

### Core RDBMS Capabilities
- ✅ **Data Types**: INT, STRING, FLOAT
- ✅ **Constraints**: PRIMARY KEY, UNIQUE, NOT NULL
- ✅ **CRUD Operations**: CREATE, INSERT, SELECT, UPDATE, DELETE
- ✅ **Indexing**: Automatic indexing on primary and unique keys
- ✅ **WHERE Clauses**: =, >, <, >=, <=, !=
- ✅ **JOIN**: Inner joins between tables
- ✅ **Concurrency**: Thread-safe with mutex locks
- ✅ **Persistence**: Auto-save to disk (minidb.json)

### Interfaces
- **REPL Mode**: Interactive SQL command-line interface
- **Web Server**: RESTful API with web UI demo
- **SQL Parser**: Custom SQL-like query language

## Usage

### REPL Mode
```bash
go run .
```

Example commands:
```sql
CREATE TABLE users (id INT PRIMARY KEY, name STRING NOT NULL, age INT)
INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30)
SELECT * FROM users
SELECT * FROM users WHERE age > 25
UPDATE users SET age = 31 WHERE id = 1
DELETE FROM users WHERE id = 1
exit
```

### Web Server Mode
```bash
go run . server
```

### Web Interfaces
- `GET /` or `GET /console` - SQL Console (Interactive query interface)
- `GET /tasks` - Task Manager Demo (CRUD application)

### API Endpoints
- `GET /api/tasks` - List all tasks
- `POST /api/tasks` - Create task (JSON body: {id, title, description, status, priority})
- `PUT /api/tasks/{id}` - Update task (JSON body: {status, ...})
- `DELETE /api/tasks/{id}` - Delete task
- `POST /api/query` - Execute SQL query (JSON body: {query})

## Architecture

### Components
- **main.go** - Entry point and REPL
- **database.go** - Database engine with concurrency control
- **table.go** - Table structure with indexing
- **sql-parser.go** - SQL query parser
- **persistence.go** - JSON-based persistence layer
- **webserver.go** - HTTP server and web UI

### Data Storage
- In-memory storage with automatic persistence to `minidb.json`
- Data survives restarts
- Thread-safe concurrent access

## Performance Features
- Index-based lookups for primary/unique keys
- Optimized WHERE clause evaluation
- Efficient row updates with index maintenance

## Demo Script
See `demo.sql` for example queries with 2 tables and JOIN operations.
