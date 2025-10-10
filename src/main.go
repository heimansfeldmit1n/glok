package main

import (
	// Package for SQL database interactions
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
	id := getYesterdayMidnightUnix()
	tasks := getTasks(id)
	day := getDay(id)
	return showWork(day, tasks)
}

func showWorkToday() string {
	id := getTodayMidnightUnix()
	tasks := getTasks(id)
	day := getDay(id)
	return showWork(day, tasks)
}

func showWorkWeek() string {
	now := time.Now()
	weekday := int(now.Weekday())
	var days []Day
	if weekday == 0 {
		weekday = 7
	}

	for i := weekday; i >= 1; i-- {
		day := now.AddDate(0, 0, -i+1)
		dayId := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()).Unix()

		d := getDay(dayId)
		if d.start.Unix() == -62135596800 {
			d.start = time.Unix(dayId, 0)
		}
		days = append(days, d)
	}

	return gatherWorkWeek(days)
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
