package main

import (
	"fmt"
	"math"
	"time"
)

import "hammy"

// Returns data for tests
type TestDataReader struct {
}

func (tr *TestDataReader) Read(hostKey string, itemKey string, from uint64, to uint64) (data []hammy.IncomingValueData, err error) {
	if hostKey != "__test" {
		panic(fmt.Sprintf("Requested test data for key %#v", hostKey))
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
		case "dmagefunc":
			n := to - from + 1
			a := 0.05
			data = make([]hammy.IncomingValueData, n)
			var i uint64
			for i = 0; i < n; i++ {
				data[i].Timestamp = from + i
				t := float64(from + i)
				data[i].Value =
					1.5*math.Sin(51*a*t) +
					-1.4*math.Sin(21*a*t) +
					1.3*math.Sin(11*a*t) +
					-1.2*math.Sin(9*a*t) +
					1.1*math.Sin(2*a*t) +
					5*math.Sin(1.73205*a*t) +
					 math.Sin(a*t) +
					7*math.Sin(0.707106*a*t) +
					 math.Sin(0.6*a*t) +
					10*math.Sin(0.1*a*t) +
					11*math.Sin(0.223606*a*t) +
					7*math.Sin(0.173205*a*t) +
					5*math.Sin(0.141421*a*t) +
					20*math.Sin(0.02*a*t) +
					5*math.Sin(0.03*a*t) +
					20*math.Sin(0.025*a*t) +
					11*math.Sin(0.0158770*a*t) +
					17*math.Sin(0.00294565*a*t) +
					19*math.Sin(0.000977518*a*t) +
					11*math.Sin(0.0000305185*a*t)
			}
		default:
			err = fmt.Errorf("Not Found")
			return
	}

	return
}

func GenTestState()(ans hammy.StateKeeperAnswer) {
	now := uint64(time.Now().Unix())

	ans.State = make(hammy.State)
	ans.State["sin"] = hammy.StateElem{
		LastUpdate: now,
		Value: 1.0,
	}
	ans.State["dmagefunc"] = hammy.StateElem{
		LastUpdate: now,
		Value: 1.0,
	}
	ans.State["Hello"] = hammy.StateElem{
		LastUpdate: now,
		Value: "world!",
	}
	return
}