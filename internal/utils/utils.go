package utils

import (
	"log"
	"math"
	"strconv"
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
