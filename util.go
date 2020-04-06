package main

import "time"

func getCurrentGameDay(date time.Time) string {
	// game day rolls over at 9am
	if date.Hour() < 9 {
		return date.Add(-24 * time.Hour).Format(basicDateFormat)
	}

	return date.Format(basicDateFormat)
}
