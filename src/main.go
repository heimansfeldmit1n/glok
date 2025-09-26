package main

import (
	"database/sql" // Package for SQL database interactions
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type Task struct {
	description string
	minutes     int
	hours       int
	timestamp   time.Time
}

type Day struct {
	id    int64
	start time.Time
	stop  time.Time
	tasks []Task
}

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

func getWorkTime(duration string) (int, int) {
	var h, m int

	s1 := strings.Split(duration, "h")
	h, err := strconv.Atoi(s1[0])
	if err != nil {
		os.Exit(1)
	}
	s2 := strings.Split(s1[1], "m")

	m, err = strconv.Atoi(s2[0])
	if err != nil {
		os.Exit(1)
	}

	return h, m
}

func getTodayMidnightUnix() int64 {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return midnight.Unix()
}

func getCurrentDay() Day {
	id := getTodayMidnightUnix()

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

func updateCurrentDay(cd Day) bool {
	_, err := DB.Exec(
		"UPDATE Day SET start = ?, stop = ? WHERE id = ?",
		cd.start, cd.stop, cd.id,
	)
	if err != nil {
		return false
	}
	return true
}

func putCurrentDay(cd Day) bool {

	id := getTodayMidnightUnix()
	_, err := DB.Exec("INSERT INTO Day (id, start, stop) VALUES (?, ?, ?)", id, cd.start, cd.stop)

	if err != nil {
		fmt.Print(err)
		return false
	}
	return true
}

func putTask(t Task) bool {
	id := getTodayMidnightUnix()

	_, err := DB.Exec("INSERT INTO Task (day_id, description, minutes, hours, timestamp) VALUES (?, ?, ?, ?, ?)", id, t.description, t.minutes, t.hours, time.Now())

	if err != nil {
		fmt.Print(err)
		return false
	}
	return true
}

func getTodaysTasks() []Task {
	id := getTodayMidnightUnix()
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

func showWorkToday() string {
	var res string

	day := getCurrentDay()

	clockedIn := day.start.Unix()
	clockedOut := day.stop.Unix()
	workTime := clockedOut - clockedIn

	res += fmt.Sprintf("Clocked in at: %s\n", day.start.Format("15:04"))
	res += fmt.Sprintf("Clocked out at: %s\n", day.stop.Format("15:04"))
	res += fmt.Sprintf("Total work time: %dh%dm\n", workTime/3600, (workTime%3600)/60)
	res += "----------------------------------\n"
	res += "Tasks:\n"

	tasks := getTodaysTasks()

	for _, t := range tasks {
		res += fmt.Sprintf("%dh%dm - %s\n", t.hours, t.minutes, t.description)
	}
	return res
}

func main() {
	initDB()
	//var version = 0.1
	start := flag.Bool("start", false, "Clock in")
	stop := flag.Bool("stop", false, "Clock out")
	versionFlag := flag.Bool("version", false, "Return Clock version")
	duration := flag.String("time", "", "How long did you take for a task 0h0m")
	flag.StringVar(duration, "t", "", "How long did you take for a task 0h0m")
	description := flag.String("description", "", "What was your task")
	flag.StringVar(description, "d", "", "What was your task")

	today := flag.Bool("today", false, "What did i do today")

	// Parse the command-line flags
	flag.Parse()

	if !(*start || *stop || *versionFlag || *today) && (len(*description) == 0 || len(*duration) == 0) {
		fmt.Println("Missing flags\n")
		os.Exit(1)
	}

	currentDay := getCurrentDay()

	if *start {
		currentDay.start = time.Now()
	} else if *stop {
		currentDay.stop = time.Now()
	} else if *today {
		str := showWorkToday()
		fmt.Println(str)
	} else {
		hur, min := getWorkTime(*duration)

		task := Task{description: *description, minutes: min, hours: hur, timestamp: time.Now()}

		putTask(task)
	}

	if currentDay.id != 0 {
		updateCurrentDay(currentDay)
	} else {
		putCurrentDay(currentDay)
	}

}
