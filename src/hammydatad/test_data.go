package main

import (
	"fmt"
	"math"
)

import "hammy"

// Returns data for tests
type TestDataReader struct {
}

func (tr *TestDataReader) Read(objKey string, itemKey string, from uint64, to uint64) (data []hammy.IncomingValueData, err error) {
	if objKey != "__test" {
		panic(fmt.Sprintf("Requested test data for key %#v", objKey))
	}

	switch itemKey {
		case "sin":
			n := to - from + 1
			data = make([]hammy.IncomingValueData, n)
			var i uint64
			for i = 0; i < n; i++ {
				data[i].Timestamp = from + i
				data[i].Value = math.Sin(float64(from + i) / 100.0 * math.Pi)
			}
		default:
			err = fmt.Errorf("Not Found")
			return
	}

	return
}