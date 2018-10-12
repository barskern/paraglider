package main

import (
	"bytes"
	"strconv"
	"time"
)

// ISO8601Duration is wrapper for `time.Duration` which implements
// `MarshalText` to render the duration as a ISO8601 compilant duration
type ISO8601Duration time.Duration

// MarshalText converts a duration into a ISO8601 compliant duration
//
// It will use the least amout of neccessary terms, so if the duration is 5
// seconds, this function will return `PT5S`.
//
// The error is always `nil`. The only way this function can fail is if the
// underlaying `bytes.Buffer` panics, which can happen if it becomes to large.
func (t ISO8601Duration) MarshalText() ([]byte, error) {
	d := time.Duration(t)

	days := int(d.Hours() / 24)

	// Two constant-sized arrays which contain the postfix and the value of a
	// ISO8601 duration element
	//
	// NB! These are written to the buffer IN ORDER, hence it is important that
	// the elements are ordered in the way same way they are supposed to be
	// ordered in the resulting byte array
	//
	// `[...]` makes sure the compiler realizes that these are constant sized
	// arrays so that they are allocated on the stack
	var dates = [...]struct {
		postfix rune
		value   int
	}{
		{'Y', int(days/365) % 365},
		// TODO take a stance on the use of months because a month does not
		// have a uniform size (can be 28,29,30,31). This results in displaying
		// the wrong days based on year, e.g. the time that should be P1Y1D
		// turns into P1Y11M24D if months are defined as 31 days (with the
		// current implementation)
		//{'M', int(days/31) % 12},
		{'D', days % 365},
	}
	var times = [...]struct {
		postfix rune
		value   int
	}{
		{'H', int(d.Hours()) % 24},
		{'M', int(d.Minutes()) % 60},
		{'S', int(d.Seconds()) % 60},
	}

	// Buffer which contains the resulting ISO8601 date
	var buffer bytes.Buffer

	// ## Write the date part of the ISO8601 duration ##

	// Tracks if any elements have been written to the buffer (to determine if
	// an empty '0S' should be prepended)
	hasElements := false
	buffer.WriteRune('P')
	for _, s := range dates {
		// Only write values which are bigger than 0
		if s.value > 0 {
			buffer.WriteString(strconv.Itoa(s.value))
			buffer.WriteRune(s.postfix)
			hasElements = true
		}
	}

	// ## Write the time part of the ISO8601 duration (if present) ##

	// Tracks wether any elements describing time have been added (to determine
	// if a 'T' is necessary)
	hasTimes := false
	for i, s := range times {
		// Only write values which are bigger than 0, BUT always include the
		// last element if no elements are present
		if s.value > 0 || (i+1 == len(times) && !hasElements) {
			// Prepends 'T' if there are no previous elements which describe
			// time
			if !hasTimes {
				buffer.WriteRune('T')
				hasTimes = true
			}
			buffer.WriteString(strconv.Itoa(s.value))
			buffer.WriteRune(s.postfix)
			hasElements = true
		}
	}
	return buffer.Bytes(), nil
}
