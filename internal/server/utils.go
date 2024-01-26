package server

import (
	"log"
	"strconv"
)

func mustParseInt64(valueString string) int64 {
	value, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		log.Fatalf("mustParseInt64 Error: %s", err.Error())
	}
	return value
}

func mustParseFloat64(valueString string) float64 {
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		log.Fatalf("mustParseFloat64 Error: %s", err.Error())
	}
	return value
}
