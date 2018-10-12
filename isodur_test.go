package main

import (
	"testing"
)

func TestISO8601DurationTimes(t *testing.T) {
	var tests = [...]struct {
		secs int
		expt string
	}{
		{0, "PT0S"},
		{1, "PT1S"},
		{60, "PT1M"},
		{125, "PT2M5S"},
		{1000, "PT16M40S"},
		{2000, "PT33M20S"},
		{18000, "PT5H"},
		{86399, "PT23H59M59S"},
	}

	for _, v := range tests {
		// Convert seconds to nanoseconds (which is the inner type of
		// `time.Duration`)
		dur := ISO8601Duration(v.secs * 1000000000)
		res, _ := dur.MarshalText()
		if string(res) != v.expt {
			t.Errorf("expected '%s' but got '%s'", v.expt, res)
		}
	}
}

func TestISO8601DurationDates(t *testing.T) {
	var tests = [...]struct {
		days int
		expt string
	}{
		{0, "PT0S"},
		{1, "P1D"},
		{5, "P5D"},
		{30, "P30D"},
		{365, "P1Y"},
		{366, "P1Y1D"},
		{600, "P1Y235D"},
	}

	for _, v := range tests {
		// Convert seconds to nanoseconds (which is the inner type of
		// `time.Duration`)
		dur := ISO8601Duration(v.days * 24 * 60 * 60 * 1000000000)
		res, _ := dur.MarshalText()
		if string(res) != v.expt {
			t.Errorf("expected '%s' but got '%s'", v.expt, res)
		}
	}
}
