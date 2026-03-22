package server

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	startTime, err := time.Parse(timeFormat, dstart)
	if err != nil {
		return "", err
	}

	// since we only work with dates, we flush time to 00:00:00
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	repeatData := strings.Split(repeat, " ")

	switch repeatData[0] {
	case "d", "y":
		return dayYearFunc(today, startTime, repeatData)
	case "w":
		return weekFunc(today, startTime, repeatData)
	case "m":
		return monthFunc(today, startTime, repeatData)
	default:
		return "", fmt.Errorf("invalid repeat format")
	}
}

func dayYearFunc(now, start time.Time, repeatData []string) (string, error) {
	if repeatData[0] == "y" && len(repeatData) == 1 {
		start = start.AddDate(1, 0, 0)
		for start.Before(now) {
			start = start.AddDate(1, 0, 0)
		}
		return start.Format(timeFormat), nil
	}

	if len(repeatData) != 2 {
		return "", fmt.Errorf("invalid repeat format")
	}

	addValue, err := strconv.Atoi(repeatData[1])
	if err != nil {
		return "", fmt.Errorf("invalid repeat value")
	}
	if addValue == 1 {
		if start.Before(now) {
			now = now.AddDate(0, 0, 1)
			return now.Format(timeFormat), nil
		}
		start = start.AddDate(0, 0, 1)
		return start.Format(timeFormat), nil
	}

	if addValue <= 0 {
		return "", fmt.Errorf("repeat value must be positive")
	}
	if addValue > 400 {
		return "", fmt.Errorf("repeat value is too large")
	}

	start = start.AddDate(0, 0, addValue)
	for start.Before(now) {
		start = start.AddDate(0, 0, addValue)
	}
	return start.Format(timeFormat), nil
}

func weekFunc(now, start time.Time, repeatData []string) (string, error) {
	if len(repeatData) != 2 {
		return "", fmt.Errorf("invalid repeat format")
	}

	repeatDaysStr := strings.Split(repeatData[1], ",")

	repeatDaysInt := make([]int, len(repeatDaysStr))

	for i, v := range repeatDaysStr {
		value, err := strconv.Atoi(v)

		if err != nil || value < 0 || value > 7 {
			return "", fmt.Errorf("invalid repeat value")
		}
		// sunday in time package corresponds to 0, not 7
		if value == 7 {
			value = 0
		}

		repeatDaysInt[i] = value
	}

	start = start.AddDate(0, 0, 1)
	for start.Before(now) || start.Equal(now) || !slices.Contains(repeatDaysInt, int(start.Weekday())) {
		start = start.AddDate(0, 0, 1)
	}

	return start.Format(timeFormat), nil
}

func monthFunc(now, start time.Time, repeatData []string) (string, error) {
	var monthDataStr []string
	var monthDataInt []int
	var checkMonth bool

	switch len(repeatData) {
	case 2:
		checkMonth = false

	case 3:
		checkMonth = true

		monthDataStr = strings.Split(repeatData[2], ",")
		monthDataInt = make([]int, len(monthDataStr))

		for i, v := range monthDataStr {
			value, err := strconv.Atoi(v)
			if err != nil || value < 1 || value > 12 {
				return "", fmt.Errorf("invalid month value")
			}
			monthDataInt[i] = value
		}

	default:
		return "", fmt.Errorf("invalid repeat format")
	}

	dayDataStr := strings.Split(repeatData[1], ",")
	dayDataInt := make([]int, len(dayDataStr))

	for i, v := range dayDataStr {
		value, err := strconv.Atoi(v)
		if err != nil || value < -2 || value == 0 || value > 31 {
			return "", fmt.Errorf("invalid day value")
		}
		dayDataInt[i] = value
	}

	start = start.AddDate(0, 0, 1)

	// cursed logical equation but it works. can be simplified TBD
	for start.Before(now) || start.Equal(now) || !((!checkMonth || slices.Contains(monthDataInt, int(start.Month()))) && isDayInMonth(start, dayDataInt)) {
		// can be optimized by adding months in some cases TBD
		start = start.AddDate(0, 0, 1)
	}
	return start.Format(timeFormat), nil
}

// slices.Contains would suffice but considering we can have negative values, I had to do something like this
func isDayInMonth(date time.Time, dayData []int) bool {
	for _, v := range dayData {
		if v < 0 && date.AddDate(0, 0, -v).Day() == 1 {
			return true
		}
		if v > 0 && date.Day() == v {
			return true
		}
	}
	return false
}
