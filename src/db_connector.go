package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// globals
var DB *sql.DB

func initDB() {
	var err error
	dbPath, _ := os.UserHomeDir()
	DB, err = sql.Open("sqlite3", dbPath+"/glok.db")
	if err != nil {
		log.Fatal(err)
	}

	// SQL statement to create the todos table if it doesn't exist
	dayStmt := `
    CREATE TABLE IF NOT EXISTS Day (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        start DATETIME,
        stop DATETIME
    );`

	// SQL statement to create the Task table
	taskStmt := `
    CREATE TABLE IF NOT EXISTS Task (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        day_id INTEGER NOT NULL,
        description TEXT,
        minutes INTEGER,
        hours INTEGER,
        timestamp DATETIME,
        FOREIGN KEY (day_id) REFERENCES Day(id) ON DELETE CASCADE
    );`
	_, err = DB.Exec(dayStmt)
	if err != nil {
		log.Fatalf("Error creating table: %q: %s\n", err, dayStmt) // Log an error if table creation fails
	}
	_, err = DB.Exec(taskStmt)
	if err != nil {
		log.Fatalf("Error creating table: %q: %s\n", err, taskStmt) // Logkan error if table creation fails
	}
}

func getDay(id int64) Day {
	res, err := DB.Query("Select * From Day Where id = ?", id)

	if err != nil {
		fmt.Println(err)
	}
	defer res.Close()
	var day Day
	if res.Next() {
		err = res.Scan(&day.id, &day.start, &day.stop)
	}

	if err != nil {
		fmt.Println(err)
		return Day{}
	}
	return day
}

func writeCurrentDay(cd Day) bool {

	_, err := DB.Exec("INSERT INTO Day (id, start, stop) VALUES (?, ?, ?)", cd.id, cd.start, cd.stop)

	if err != nil {
		fmt.Print(err)
		return false
	}
	return true
}

func writeTask(t Task, id int64) bool {

	_, err := DB.Exec("INSERT INTO Task (day_id, description, minutes, hours, timestamp) VALUES (?, ?, ?, ?, ?)", id, t.description, t.minutes, t.hours, time.Now())

	if err != nil {
		fmt.Print(err)
		return false
	}
	return true
}

func getTasks(id int64) []Task {
	var tasks []Task

	rows, err := DB.Query("Select description, minutes, hours, timestamp From Task Where day_id = ?", id)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rows.Close()
	for rows.Next() {
		var task Task

		if err := rows.Scan(&task.description, &task.minutes, &task.hours, &task.timestamp); err != nil {
			fmt.Println("Scan failed:", err)
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks
}

func updateCurrentDay(cd Day) bool {
	_, err := DB.Exec(
		"UPDATE Day SET start = ?, stop = ? WHERE id = ?",
		cd.start, cd.stop, cd.id,
	)
	if err != nil {
		os.Exit(1)
	}
	return true
}
