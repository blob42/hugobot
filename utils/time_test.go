package utils

import (
	"testing"
	"time"
)

func TestGetAllThursdays(t *testing.T) {
	tt, err := time.Parse("2006-01-02", "2017-12-07")
	if err != nil {
		t.Error(err)
	}

	dates := GetAllThursdays(tt, time.Now())

	if dates[0] != NextThursday(tt) {
		t.Error("starting date")
	}

	t.Log(NextThursday(time.Now()))
	t.Log(dates[len(dates)-1])

	if dates[len(dates)-1] != NextThursday(time.Now()) {
		t.Error("end date")
	}
}
