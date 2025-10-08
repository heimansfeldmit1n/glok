package main

import (
	"database/sql" // Package for SQL database interactions
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

func getYesterdayMidnightUnix() int64 {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1) // subtract 1 day
	midnight := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
	return midnight.Unix()
}

func showYesterdayWork() string {
	return showWork(getDay(getYesterdayMidnightUnix()))
}

func showWorkToday() string {
	return showWork(getDay(getTodayMidnightUnix()))
}

func showWork(day Day) string {
	var res string

	clockedIn := day.start.Unix()

	clockedOut := day.stop.Unix()

	if clockedOut == -62135596800 {
		clockedOut = time.Now().Unix()
	}

	workTime := clockedOut - clockedIn

	res += fmt.Sprintf("Clocked in at: %s\n", day.start.Format("15:04"))
	res += fmt.Sprintf("Clocked out at: %s\n", day.stop.Format("15:04"))
	res += fmt.Sprintf("Total work time: %dh%dm\n", workTime/3600, (workTime%3600)/60)
	res += "----------------------------------\n"
	res += "Tasks:\n"

	tasks := getTasks(day.id)

	for _, t := range tasks {
		res += fmt.Sprintf("%dh%dm - %s\n", t.hours, t.minutes, t.description)
	}
	return res
}

func showWorkWeek() string {
	var res string
	res += "Hours worked this week:\n"
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -weekday+1)
	totalWeek := 0.0

	for i := 0; i < 7; i++ {
		day := monday.AddDate(0, 0, i)
		dayID := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()).Unix()

		var start, stop sql.NullTime

		err := DB.QueryRow("Select start, stop From Day Where id = ?", dayID).Scan(&start, &stop)

		if err != nil {
			if err == sql.ErrNoRows {
				res += fmt.Sprintf("%s: 0h0m\n", day.Format("Mon 02"))
				continue
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if start.Valid && stop.Valid {
			clockedIn := start.Time.Unix()
			clockedOut := stop.Time.Unix()
			workTime := clockedOut - clockedIn
			hours := workTime / 3600
			minutes := (workTime % 3600) / 60
			totalWeek += float64(hours) + float64(minutes)/60.0
			res += fmt.Sprintf("%s: %dh%dm\n", day.Format("Mon 02"), hours, minutes)
		} else {
			res += fmt.Sprintf("%s: 0h0m\n", day.Format("Mon 02"))
		}
	}
	res += "----------------------------------\n"
	res += fmt.Sprintf("Total hours worked this week: %.1fh\n", totalWeek)
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

	yesterday := flag.Bool("yesterday", false, "What did i do yesterday")
	today := flag.Bool("today", false, "What did i do today")
	week := flag.Bool("week", false, "Hours per day worked this week")

	flag.Parse()

	if !(*start || *stop || *versionFlag || *today || *week || *yesterday) && (len(*description) == 0 || len(*duration) == 0) {
		fmt.Println("Missing flags")
		os.Exit(1)
	}

	currentDay := getDay(getTodayMidnightUnix())

	if *start {
		currentDay.start = time.Now()
	} else if *stop {
		currentDay.stop = time.Now()
	} else if *today {
		str := showWorkToday()
		fmt.Println(str)
	} else if *week {
		str := showWorkWeek()
		fmt.Println(str)
	} else if *yesterday {
		str := showYesterdayWork()
		fmt.Println(str)
	} else {
		hur, min := getWorkTime(*duration)

		task := Task{description: *description, minutes: min, hours: hur, timestamp: time.Now()}

		writeTask(task, getTodayMidnightUnix())
	}

	if currentDay.id != 0 {
		updateCurrentDay(currentDay)
	} else {
		currentDay.id = getTodayMidnightUnix()
		writeCurrentDay(currentDay)
	}

}
