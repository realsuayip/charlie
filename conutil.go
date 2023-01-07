package main

import (
	"fmt"
	"strings"
	"time"
)

type ArbitraryData = map[string]interface{}

const day = time.Minute * 60 * 24

func duration(d time.Duration) string {
	// Format duration as string, also pretty formatting days.
	if d < day {
		return d.String()
	}
	var b strings.Builder
	days := d / day
	d -= days * day

	ds := d.String()
	if d == 0 {
		ds = ""
	}
	_, _ = fmt.Fprintf(&b, "%ddays%s", days, ds)
	return b.String()
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minDate(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func newDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
