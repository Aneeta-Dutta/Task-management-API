package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
	//_ "github.com/mattn/go-sqlite3"
)

// Task represents a task in the database
type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
}

var db *sql.DB

func setupDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		return nil, err
	}

	// Create the tasks table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		Title TEXT NOT NULL,
		Description TEXT,
		DueDate TEXT,
		Status TEXT
	)`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createTask(c *gin.Context) {
	var task Task
	if err := c.BindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	stmt, err := db.Prepare("INSERT INTO tasks (Title, Description, DueDate, Status) VALUES (?, ?, ?, ?)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(task.Title, task.Description, task.DueDate, task.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	task.ID, _ = result.LastInsertId()
	c.JSON(http.StatusOK, task)
}

func getTask(c *gin.Context) {
	taskID := c.Param("id")
	var task Task

	err := db.QueryRow("SELECT ID, Title, Description, DueDate, Status FROM tasks WHERE ID = ?", taskID).Scan(
		&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func updateTask(c *gin.Context) {
	taskID := c.Param("id")
	var task Task
	if err := c.BindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	stmt, err := db.Prepare("UPDATE tasks SET Title=?, Description=?, DueDate=?, Status=? WHERE ID=?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(task.Title, task.Description, task.DueDate, task.Status, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	taskID := c.Param("id")

	_, err := db.Exec("DELETE FROM tasks WHERE ID=?", taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func listTasks(c *gin.Context) {
	rows, err := db.Query("SELECT ID, Title, Description, DueDate, Status FROM tasks")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}

func main() {
	fmt.Println("Starting Task Management API...")
	r := gin.Default()

	var err error
	db, err = setupDatabase()
	if err != nil {
		log.Fatal("Error setting up the database:", err)
	}
	defer db.Close()

	r.POST("/tasks", createTask)
	r.GET("/tasks/:id", getTask)
	r.PUT("/tasks/:id", updateTask)
	r.DELETE("/tasks/:id", deleteTask)
	r.GET("/tasks", listTasks)

	r.Run(":8080")
}
