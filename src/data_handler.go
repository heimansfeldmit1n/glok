package main

import (
	"fmt"
	"time"
)

func showWork(day Day, tasks []Task) string {
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

	for _, t := range tasks {
		res += fmt.Sprintf("%dh%dm - %s\n", t.hours, t.minutes, t.description)
	}
	return res
}

func gatherWorkWeek(days []Day) string {
	var res string
	var totalWeek float32
	res += "Hours worked this week:\n"
	for _, day := range days {
		start := day.start
		stop := day.stop
		clockedIn := start.Unix()
		clockedOut := stop.Unix()
		workTime := clockedOut - clockedIn
		if workTime < 0 {
			workTime = 0
		}
		hours := workTime / 3600
		minutes := (workTime % 3600) / 60
		totalWeek += float32(hours) + float32(minutes)/60.0
		res += fmt.Sprintf("%s: %dh%dm\n", start.Format("Mon 02"), hours, minutes)
	}
	res += "----------------------------------\n"
	res += fmt.Sprintf("Total hours worked this week: %.1fh\n", totalWeek)
	return res
}
