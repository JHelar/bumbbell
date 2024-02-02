package utils

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
)

func MustParseInt64(valueString string) int64 {
	value, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		log.Fatalf("MustParseInt64 Error: %s", err.Error())
	}
	return value
}

func MustParseFloat64(valueString string) float64 {
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		log.Fatalf("MustParseFloat64 Error: %s", err.Error())
	}
	return value
}

// Percent - calculate what %[number1] of [number2] is.
// ex. 25% of 200 is 50
func Percent(percent int, all int) float64 {
	return ((float64(all) * float64(percent)) / float64(100))
}

// Change - calculate the percent increase/decrease from two numbers.
// ex. 60 is a 200.0% increase from 20
func Change(before int, after int) float64 {
	res := 100 * ((float64(after) - float64(before)) / float64(before))
	if res == math.Inf(1) {
		return 100
	} else if res == math.Inf(-1) {
		return -100
	}
	return res
}

// PercentOf - calculate what percent [number1] is of [number2].
// ex. 300 is 12.5% of 2400
func PercentOf(part int, total int) float64 {
	return (float64(part) * float64(100)) / float64(total)
}

func FmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour

	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%02dm %02ds", m, s)
	}
	return fmt.Sprintf("%02ds", s)
}
