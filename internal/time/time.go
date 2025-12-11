package time

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var re regexp.Regexp = *regexp.MustCompile(`^\d+[smh]$`)

func GetTimeInterval(interval string) (time.Duration, error) {
	if !re.Match([]byte(interval)) {
		return 0, fmt.Errorf("interval string does not match format (i.e 1s, 3m, 5h): %v", interval)
	}

	timeValStr, lastChar := interval[:len(interval)-1], interval[len(interval)-1]
	timeVal, err := strconv.ParseInt(timeValStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("util/util.go: unable to ParseInt(timevalStr)")
	}

	switch lastChar {
	case 's':
		return time.Duration(timeVal) * time.Second, nil
	case 'm':
		return time.Duration(timeVal) * time.Minute, nil
	case 'h':
		return time.Duration(timeVal) * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid time unit: %v", lastChar)
	}

}
