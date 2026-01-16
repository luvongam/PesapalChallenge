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
	initializeDB()

	http.HandleFunc("/", handleHome)
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
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>Task Manager - SimpleDB Demo</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1000px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .task-form { background: #f5f5f5; padding: 20px; border-radius: 5px; margin-bottom: 30px; }
        .task-form input, .task-form textarea, .task-form select { 
            width: 100%; padding: 8px; margin: 5px 0 15px 0; border: 1px solid #ddd; border-radius: 3px;
        }
        .task-form button { background: #4CAF50; color: white; padding: 10px 20px; border: none; border-radius: 3px; cursor: pointer; }
        .task-form button:hover { background: #45a049; }
        .task-list { margin-top: 20px; }
        .task-item { background: white; border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .task-item h3 { margin-top: 0; color: #333; }
        .task-item .status { display: inline-block; padding: 3px 10px; border-radius: 3px; font-size: 12px; }
        .status.pending { background: #fff3cd; color: #856404; }
        .status.in-progress { background: #cce5ff; color: #004085; }
        .status.completed { background: #d4edda; color: #155724; }
        .task-item .actions { margin-top: 10px; }
        .task-item button { margin-right: 5px; padding: 5px 10px; border: none; border-radius: 3px; cursor: pointer; }
        .btn-update { background: #007bff; color: white; }
        .btn-delete { background: #dc3545; color: white; }
        .priority { color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <h1>📋 Task Manager</h1>
    <p>Simple CRUD demo powered by SimpleDB RDBMS</p>

    <div class="task-form">
        <h2>Add New Task</h2>
        <input type="text" id="title" placeholder="Task Title" required>
        <textarea id="description" placeholder="Description" rows="3"></textarea>
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

    <div class="task-list" id="taskList">
        <h2>All Tasks</h2>
    </div>

    <script>
        let nextId = 1;

        async function loadTasks() {
            const response = await fetch('/api/tasks');
            const tasks = await response.json();
            
            const taskList = document.getElementById('taskList');
            taskList.innerHTML = '<h2>All Tasks</h2>';
            
            if (tasks.length === 0) {
                taskList.innerHTML += '<p>No tasks yet. Add one above!</p>';
                return;
            }

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
                body: JSON.stringify({
                    id: nextId++,
                    title,
                    description,
                    status,
                    priority
                })
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

            if (response.ok) {
                loadTasks();
            }
        }

        async function deleteTask(id) {
            if (!confirm('Delete this task?')) return;

            const response = await fetch(` + "`" + `/api/tasks/${id}` + "`" + `, {
                method: 'DELETE'
            });

            if (response.ok) {
                loadTasks();
            }
        }

        loadTasks();
    </script>
</body>
</html>`

	tmplObj, _ := template.New("home").Parse(tmpl)
	tmplObj.Execute(w, nil)
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
			json.NewEncoder(w).Encode([]Row{})
		} else {
			json.NewEncoder(w).Encode(result.Rows)
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
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
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

		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})

	case "DELETE":
		query := fmt.Sprintf("DELETE FROM tasks WHERE id = %d", id)
		_, err := globalDB.Execute(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	}
}
