package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var globalDB *Database

func startWebServer() {
	globalDB = NewDatabase()
	err := globalDB.Load()
	if err != nil {
		return
	}
	initializeDB()

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/console", handleConsole)
	http.HandleFunc("/tasks", handleTaskManager)
	http.HandleFunc("/api/query", handleQuery)
	http.HandleFunc("/api/tasks", handleTasks)
	http.HandleFunc("/api/tasks/", handleTaskByID)

	fmt.Println("Web server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initializeDB() {
	query := `CREATE TABLE tasks (
		id INT PRIMARY KEY,
		title STRING NOT NULL,
		description STRING,
		status STRING,
		priority INT
	)`
	log.Printf("Creating table with query: %s", query)
	_, err := globalDB.Execute(query)

	if err != nil {
		log.Printf("Error creating table: %v", err)
	} else {
		log.Printf("Table created successfully")
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/console", http.StatusFound)
}

func handleConsole(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SimpleDB Console</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; }
        .container { max-width: 1400px; margin: 0 auto; padding: 20px; }
        .header { background: rgba(255,255,255,0.95); backdrop-filter: blur(10px); border-radius: 20px; padding: 30px; margin-bottom: 20px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); }
        .header h1 { font-size: clamp(24px, 5vw, 36px); background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; margin-bottom: 10px; }
        .header p { color: #666; font-size: clamp(12px, 2vw, 14px); }
        .nav { display: flex; gap: 10px; margin-bottom: 20px; flex-wrap: wrap; }
        .nav a { background: rgba(255,255,255,0.9); color: #667eea; text-decoration: none; padding: 12px 24px; border-radius: 12px; font-weight: 600; transition: all 0.3s; box-shadow: 0 4px 15px rgba(0,0,0,0.1); }
        .nav a:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(0,0,0,0.15); }
        .nav a.active { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
        .main-content { display: grid; grid-template-columns: 1fr; gap: 20px; }
        @media (min-width: 768px) { .main-content { grid-template-columns: 1fr 1fr; } }
        .panel { background: rgba(255,255,255,0.95); backdrop-filter: blur(10px); border-radius: 20px; padding: 25px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); }
        .panel h2 { color: #333; margin-bottom: 20px; font-size: clamp(18px, 3vw, 24px); }
        textarea { width: 100%; min-height: 200px; background: #f8f9fa; border: 2px solid #e9ecef; border-radius: 12px; padding: 15px; font-family: 'Courier New', monospace; font-size: 14px; resize: vertical; transition: all 0.3s; }
        textarea:focus { outline: none; border-color: #667eea; box-shadow: 0 0 0 3px rgba(102,126,234,0.1); }
        .btn-group { display: flex; gap: 10px; margin-top: 15px; flex-wrap: wrap; }
        button { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none; padding: 12px 24px; border-radius: 12px; cursor: pointer; font-weight: 600; font-size: 14px; transition: all 0.3s; box-shadow: 0 4px 15px rgba(102,126,234,0.4); }
        button:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(102,126,234,0.6); }
        button:active { transform: translateY(0); }
        .btn-secondary { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); box-shadow: 0 4px 15px rgba(245,87,108,0.4); }
        .btn-secondary:hover { box-shadow: 0 6px 20px rgba(245,87,108,0.6); }
        .results { background: #f8f9fa; border-radius: 12px; padding: 20px; min-height: 300px; max-height: 500px; overflow: auto; font-family: 'Courier New', monospace; font-size: 13px; }
        .success { color: #10b981; padding: 12px; background: #d1fae5; border-radius: 8px; margin-bottom: 15px; font-weight: 600; }
        .error { color: #ef4444; padding: 12px; background: #fee2e2; border-radius: 8px; margin-bottom: 15px; font-weight: 600; }
        table { width: 100%; border-collapse: collapse; margin-top: 15px; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        th { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 12px; text-align: left; font-weight: 600; }
        td { padding: 12px; border-bottom: 1px solid #e9ecef; }
        tr:hover { background: #f8f9fa; }
        tr:last-child td { border-bottom: none; }
        .examples { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; border-radius: 8px; margin-top: 15px; font-size: 12px; }
        .examples code { color: #d63384; background: white; padding: 2px 6px; border-radius: 4px; }
        .loading { text-align: center; color: #667eea; padding: 20px; }
        @media (max-width: 767px) { .main-content { grid-template-columns: 1fr; } .panel { padding: 20px; } }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🗄️ SimpleDB RDBMS Console</h1>
            <p>Modern SQL Database Management System</p>
        </div>
        <div class="nav">
            <a href="/console" class="active">SQL Console</a>
            <a href="/tasks">Task Manager Demo</a>
        </div>
        <div class="main-content">
            <div class="panel">
                <h2>📝 SQL Editor</h2>
                <textarea id="query" placeholder="Enter your SQL query here...

Try these examples:
CREATE TABLE users (id INT PRIMARY KEY, name STRING NOT NULL, age INT)
INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30)
SELECT * FROM users
SELECT * FROM users WHERE age > 25"></textarea>
                <div class="btn-group">
                    <button onclick="executeQuery()">▶ Execute Query</button>
                    <button class="btn-secondary" onclick="clearAll()">🗑️ Clear All</button>
                </div>
                <div class="examples">
                    <strong>💡 Quick Tips:</strong><br>
                    • Press <code>Ctrl+Enter</code> to execute<br>
                    • Supports: CREATE, INSERT, SELECT, UPDATE, DELETE<br>
                    • Data types: INT, STRING, FLOAT
                </div>
            </div>
            <div class="panel">
                <h2>📊 Query Results</h2>
                <div class="results" id="results">
                    <div style="color: #999; text-align: center; padding: 40px;">Execute a query to see results...</div>
                </div>
            </div>
        </div>
    </div>
    <script>
        async function executeQuery() {
            const query = document.getElementById('query').value.trim();
            if (!query) {
                alert('⚠️ Please enter a query');
                return;
            }

            const resultsDiv = document.getElementById('results');
            resultsDiv.innerHTML = '<div class="loading">⏳ Executing query...</div>';

            try {
                const response = await fetch('/api/query', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ query })
                });

                const data = await response.json();

                if (!response.ok) {
                    resultsDiv.innerHTML = ` + "`<div class=\"error\">❌ Error: ${data.error}</div>`" + `;
                    return;
                }

                if (data.message) {
                    resultsDiv.innerHTML = ` + "`<div class=\"success\">✅ ${data.message}</div>`" + `;
                } else if (data.rows && data.rows.length > 0) {
                    let html = ` + "`<div class=\"success\">✅ Query executed successfully (${data.rows.length} rows returned)</div>`" + `;
                    html += '<table><thead><tr>';
                    
                    const cols = data.columns || Object.keys(data.rows[0]);
                    cols.forEach(col => {
                        html += ` + "`<th>${col}</th>`" + `;
                    });
                    html += '</tr></thead><tbody>';
                    
                    data.rows.forEach(row => {
                        html += '<tr>';
                        cols.forEach(col => {
                            html += ` + "`<td>${row[col] !== null && row[col] !== undefined ? row[col] : '<em style=\"color:#999\">NULL</em>'}</td>`" + `;
                        });
                        html += '</tr>';
                    });
                    html += '</tbody></table>';
                    resultsDiv.innerHTML = html;
                } else {
                    resultsDiv.innerHTML = '<div class="success">✅ Query executed successfully (0 rows returned)</div>';
                }
            } catch (err) {
                resultsDiv.innerHTML = ` + "`<div class=\"error\">❌ Network Error: ${err.message}</div>`" + `;
            }
        }

        function clearAll() {
            document.getElementById('query').value = '';
            document.getElementById('results').innerHTML = '<div style="color: #999; text-align: center; padding: 40px;">Execute a query to see results...</div>';
        }

        document.getElementById('query').addEventListener('keydown', function(e) {
            if (e.ctrlKey && e.key === 'Enter') {
                executeQuery();
            }
        });
    </script>
</body>
</html>`

	tmplObj, _ := template.New("home").Parse(tmpl)
	err := tmplObj.Execute(w, nil)
	if err != nil {
		return
	}
}

func handleTaskManager(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Task Manager - SimpleDB Demo</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: rgba(255,255,255,0.95); backdrop-filter: blur(10px); border-radius: 20px; padding: 30px; margin-bottom: 20px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); }
        .header h1 { font-size: clamp(24px, 5vw, 36px); background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; margin-bottom: 10px; }
        .header p { color: #666; font-size: clamp(12px, 2vw, 14px); }
        .nav { display: flex; gap: 10px; margin-bottom: 20px; flex-wrap: wrap; }
        .nav a { background: rgba(255,255,255,0.9); color: #667eea; text-decoration: none; padding: 12px 24px; border-radius: 12px; font-weight: 600; transition: all 0.3s; box-shadow: 0 4px 15px rgba(0,0,0,0.1); }
        .nav a:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(0,0,0,0.15); }
        .nav a.active { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
        .panel { background: rgba(255,255,255,0.95); backdrop-filter: blur(10px); border-radius: 20px; padding: 25px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); margin-bottom: 20px; }
        .panel h2 { color: #333; margin-bottom: 20px; font-size: clamp(18px, 3vw, 24px); }
        input, textarea, select { width: 100%; padding: 12px; margin: 8px 0; background: #f8f9fa; border: 2px solid #e9ecef; border-radius: 8px; font-size: 14px; transition: all 0.3s; }
        input:focus, textarea:focus, select:focus { outline: none; border-color: #667eea; box-shadow: 0 0 0 3px rgba(102,126,234,0.1); }
        button { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none; padding: 12px 24px; border-radius: 12px; cursor: pointer; font-weight: 600; font-size: 14px; transition: all 0.3s; box-shadow: 0 4px 15px rgba(102,126,234,0.4); margin-top: 10px; }
        button:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(102,126,234,0.6); }
        .task-item { background: white; border-radius: 12px; padding: 20px; margin: 15px 0; box-shadow: 0 2px 8px rgba(0,0,0,0.1); transition: all 0.3s; }
        .task-item:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
        .task-item h3 { color: #333; margin-bottom: 10px; }
        .task-item p { color: #666; margin-bottom: 15px; }
        .status { display: inline-block; padding: 6px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; margin-right: 10px; }
        .status.pending { background: #fff3cd; color: #856404; }
        .status.in-progress { background: #cce5ff; color: #004085; }
        .status.completed { background: #d4edda; color: #155724; }
        .priority { color: #999; font-size: 14px; }
        .actions { margin-top: 15px; display: flex; gap: 10px; flex-wrap: wrap; }
        .btn-update { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); padding: 8px 16px; }
        .btn-delete { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); padding: 8px 16px; }
        .empty { text-align: center; color: #999; padding: 40px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📋 Task Manager</h1>
            <p>CRUD Demo - Powered by SimpleDB RDBMS</p>
        </div>
        <div class="nav">
            <a href="/console">SQL Console</a>
            <a href="/tasks" class="active">Task Manager Demo</a>
        </div>
        <div class="panel">
            <h2>➕ Add New Task</h2>
            <input type="text" id="title" placeholder="Task Title" required>
            <textarea id="description" placeholder="Task Description" rows="3"></textarea>
            <select id="status">
                <option value="pending">Pending</option>
                <option value="in-progress">In Progress</option>
                <option value="completed">Completed</option>
            </select>
            <select id="priority">
                <option value="1">Low Priority</option>
                <option value="2">Medium Priority</option>
                <option value="3">High Priority</option>
            </select>
            <button onclick="createTask()">Add Task</button>
        </div>
        <div class="panel">
            <h2>📝 All Tasks</h2>
            <div id="taskList"></div>
        </div>
    </div>
    <script>
        let nextId = 1;

        async function loadTasks() {
            const response = await fetch('/api/tasks');
            const tasks = await response.json();
            
            const taskList = document.getElementById('taskList');
            
            if (tasks.length === 0) {
                taskList.innerHTML = '<div class="empty">No tasks yet. Add one above!</div>';
                return;
            }

            taskList.innerHTML = '';
            tasks.forEach(task => {
                const div = document.createElement('div');
                div.className = 'task-item';
                div.innerHTML = ` + "`" + `
                    <h3>${task.title}</h3>
                    <p>${task.description || 'No description'}</p>
                    <span class="status ${task.status}">${task.status}</span>
                    <span class="priority">Priority: ${task.priority}</span>
                    <div class="actions">
                        <button class="btn-update" onclick="updateStatus(${task.id}, '${task.status}')">Change Status</button>
                        <button class="btn-delete" onclick="deleteTask(${task.id})">Delete</button>
                    </div>
                ` + "`" + `;
                taskList.appendChild(div);
                
                if (task.id >= nextId) {
                    nextId = task.id + 1;
                }
            });
        }

        async function createTask() {
            const title = document.getElementById('title').value;
            const description = document.getElementById('description').value;
            const status = document.getElementById('status').value;
            const priority = parseInt(document.getElementById('priority').value);

            if (!title) {
                alert('Title is required');
                return;
            }

            const response = await fetch('/api/tasks', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: nextId++, title, description, status, priority })
            });

            if (response.ok) {
                document.getElementById('title').value = '';
                document.getElementById('description').value = '';
                loadTasks();
            }
        }

        async function updateStatus(id, currentStatus) {
            const statuses = ['pending', 'in-progress', 'completed'];
            const currentIndex = statuses.indexOf(currentStatus);
            const newStatus = statuses[(currentIndex + 1) % statuses.length];

            const response = await fetch(` + "`" + `/api/tasks/${id}` + "`" + `, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ status: newStatus })
            });

            if (response.ok) loadTasks();
        }

        async function deleteTask(id) {
            if (!confirm('Delete this task?')) return;

            const response = await fetch(` + "`" + `/api/tasks/${id}` + "`" + `, {
                method: 'DELETE'
            });

            if (response.ok) loadTasks();
        }

        loadTasks();
    </script>
</body>
</html>`

	tmplObj, _ := template.New("tasks").Parse(tmpl)
	err := tmplObj.Execute(w, nil)
	if err != nil {
		return
	}
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		result, err := globalDB.Execute("SELECT * FROM tasks")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if result.Rows == nil {
			err := json.NewEncoder(w).Encode([]Row{})
			if err != nil {
				return
			}
		} else {
			err := json.NewEncoder(w).Encode(result.Rows)
			if err != nil {
				return
			}
		}

	case "POST":
		var task map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			log.Printf("Decode error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query := fmt.Sprintf(
			"INSERT INTO tasks (id, title, description, status, priority) VALUES (%d, '%s', '%s', '%s', %d)",
			int(task["id"].(float64)),
			task["title"],
			task["description"],
			task["status"],
			int(task["priority"].(float64)),
		)

		log.Printf("Executing query: %s", query)
		_, err := globalDB.Execute(query)
		if err != nil {
			log.Printf("Execute error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"status": "created"})
		if err != nil {
			return
		}
	}
}

func handleTaskByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	idStr := path[len("/api/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var setClauses []string
		for k, v := range updates {
			setClauses = append(setClauses, fmt.Sprintf("%s = '%v'", k, v))
		}

		query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = %d",
			strings.Join(setClauses, ", "), id)

		_, err := globalDB.Execute(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
		if err != nil {
			return
		}

	case "DELETE":
		query := fmt.Sprintf("DELETE FROM tasks WHERE id = %d", id)
		_, err := globalDB.Execute(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
		if err != nil {
			return
		}
	}
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		if err != nil {
			return
		}
		return
	}

	result, err := globalDB.Execute(req.Query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		if err != nil {
			return
		}
		return
	}

	if result.Message != "" {
		err := json.NewEncoder(w).Encode(map[string]string{"message": result.Message})
		if err != nil {
			return
		}
	} else {
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"columns": result.Columns,
			"rows":    result.Rows,
		})
		if err != nil {
			return
		}
	}
}
