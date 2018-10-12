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
		// Convert to nanoseconds
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
		{600, "P1Y7M18D"},
		{800, "P2Y2M8D"},
		{801, "P2Y2M9D"},
		{802, "P2Y2M10D"},
		{803, "P2Y2M11D"},
		{834, "P2Y3M11D"},
		{100000, "P273Y11M14D"},
	}

	for _, v := range tests {
		// Convert to nanoseconds
		dur := ISO8601Duration(v.days * 24 * 60 * 60 * 1000000000)
		res, _ := dur.MarshalText()
		if string(res) != v.expt {
			t.Errorf("expected '%s' but got '%s'", v.expt, res)
		}
	}
}

func TestISO8601DurationDatesWithTimes(t *testing.T) {
	var tests = [...]struct {
		days int
		secs int
		expt string
	}{
		{0, 0, "PT0S"},
		{1, 1, "P1DT1S"},
		{12, 12, "P12DT12S"},
		{12, 200, "P12DT3M20S"},
		{12, 60 * 60 * 24, "P13D"},
		{30, 60 * 60 * 24, "P1M"},
		{364, 60 * 60 * 24, "P1Y"},
		{365, 60 * 60 * 24, "P1Y1D"},
		{366, 60 * 60, "P1Y1DT1H"},
		{366, 60*60 + 1, "P1Y1DT1H1S"},
		{100000, 60, "P273Y11M14DT1M"},
	}

	for _, v := range tests {
		// Convert to nanoseconds
		dur := ISO8601Duration((v.secs + v.days*24*60*60) * 1000000000)
		res, _ := dur.MarshalText()
		if string(res) != v.expt {
			t.Errorf("expected '%s' but got '%s'", v.expt, res)
		}
	}
}
