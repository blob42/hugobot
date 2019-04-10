package utils

import (
	"log"
	"time"
)

func NextThursday(t time.Time) time.Time {
	weekday := t.Weekday()

	t, err := time.Parse("2006-01-02", t.Format("2006-01-02"))
	if err != nil {
		log.Println(err)
	}
	nextThursday := t

	if weekday < 4 {
		nextThursday = t.AddDate(0, 0, int(4-weekday))
	} else if weekday > 4 {
		nextThursday = t.AddDate(0, 0, int((7-weekday)+4))
	}

	return nextThursday
}

// Returns all thursdays starting from now up to the input date
func GetAllThursdays(from time.Time, to time.Time) []time.Time {
	var dates []time.Time

	//log.Printf("Parsing from %s", from)

	firstWeek := NextThursday(from)
	lastWeek := NextThursday(to)

	//log.Printf("First thursday is %s", firstWeek)

	cursorWeek := firstWeek
	for cursorWeek.Before(lastWeek) {
		dates = append(dates, cursorWeek)
		cursorWeek = cursorWeek.AddDate(0, 0, 7)
	}

	if !cursorWeek.Before(lastWeek) &&
		cursorWeek.Weekday() == time.Thursday {
		dates = append(dates, cursorWeek)
	}

	return dates
}
