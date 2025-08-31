package testing

import (
	"testing"
	"time"

	"bebop831.com/filo/util"
)

func TestGetTimeInterval(t *testing.T) {
	test_data := map[string]time.Duration{
		// Valid basics
		"1s":  time.Second,
		"10m": 10 * time.Minute,
		"2h":  2 * time.Hour,

		// Boundary numbers
		"0s":                    0,
		"999999999s":            999999999 * time.Second, // huge, may overflow
		"0001s":                 time.Second,
		"18446744073709551615s": 0, // should fail

		// Weird suffixes
		"1sec":     0,
		"2minutes": 0,
		"3hour":    0,
		"5ss":      0,
		"10hh":     0,

		// Nonnumeric junk
		"5s5":  0,
		"12m3": 0,
		"５s":   0,
		"1💥s":  0,

		// Whitespace / control chars
		" 1s":      0,
		"1s ":      0,
		"1\tm":     0,
		"1\nh":     0,
		"1\u200Bs": 0,

		// Signs and decimals
		"-10s": 0,
		"+5m":  0,
		"1.5h": 0,

		// Case traps
		"1S":   0,
		"1M":   0,
		"1H":   0,
		"10Ms": 0,

		// Empty / random junk
		"":    0,
		" ":   0,
		"abc": 0,
		"💣":   0,
		"123": 0,
		"s":   0,
	}

	for timeStr, expectedTimeVal := range test_data {
		currentTimeVal, err := util.GetTimeInterval(timeStr)
		if err != nil && expectedTimeVal != 0 {
			t.Error(err)
		}

		if currentTimeVal != expectedTimeVal {
			t.Errorf("util.GetTimeInterval(%s) != %v\n", timeStr, expectedTimeVal)
		}
	}
}
