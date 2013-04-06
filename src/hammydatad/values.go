package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"strings"
	"sort"
)

import "hammy"

type ValuesAnswer struct {
	X []uint64
	Y []interface{}
}

func (h *HttpServer) ServeValues(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	key_a := q["key"]
	if len(key_a) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	key := key_a[0]
	if key == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	host_a := q["host"]
	if len(host_a) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	host := host_a[0]
	if host == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var from, to uint64
	from_a := q["from"]
	if len(from_a) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscan(from_a[0], &from); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	to_a := q["to"]
	if len(to_a) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscan(to_a[0], &to); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if from >= to {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	aggr_a := q["aggr"]
	var aggr uint64 = 0
	if len(aggr_a) != 0 {
		if _, err := fmt.Sscan(aggr_a[0], &aggr); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}

	quantiles_a := q["quantiles"]
	var quantiles []float64
	if aggr > 0 && len(quantiles_a) != 0 {
		q_strs := strings.Split(quantiles_a[0], ",")
		for i := range q_strs {
			var q float64
			if _, err := fmt.Sscan(q_strs[i], &q); err != nil {
				http.Error(w, "Bad Request (quantiles)", http.StatusBadRequest)
				return
			}
			if q < 0 || q > 1 {
				http.Error(w, "Bad Request (quantiles)", http.StatusBadRequest)
				return
			}
			quantiles = append(quantiles, q)
		}
	}

	var dataReader hammy.DataReader
	if host == "__test" {
		dataReader = &TestDataReader{}
	} else {
		dataReader = h.DReader
	}

	data, err := dataReader.Read(host, key, from, to)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", err)
		log.Printf("Internal Server Error: %v", err)
		return
	}

	ans := new(ValuesAnswer)
	n := len(data)
	if aggr <= 0 {
		ans.X = make([]uint64, n)
		ans.Y = make([]interface{}, n)
		for i := 0; i < n; i++ {
			ans.X[i] = data[i].Timestamp
			ans.Y[i] = data[i].Value
		}
	} else {
		var prevt uint64 = data[0].Timestamp/aggr*aggr
		if len(quantiles) != 0 {
			var values []float64
			for i := 0; i < n; i++ {
				var t uint64 = data[i].Timestamp/aggr*aggr
				if t != prevt && len(values) > 0 {
					sort.Float64s(values)
					var result []float64
					result = make([]float64, len(quantiles))
					for j := 0; j < len(quantiles); j++ {
						idx := int(float64(len(values))*quantiles[j])
						if idx == len(values) {
							idx = len(values) - 1
						}
						result[j] = values[idx]
					}
					ans.X = append(ans.X, prevt)
					ans.Y = append(ans.Y, result)
					values = []float64{}
				}
				if value, ok := data[i].Value.(float64); ok {
					values = append(values, value)
				}
				prevt = t
			}
		} else {
			var count uint64 = 0
			var sum float64 = 0
			for i := 0; i < n; i++ {
				var t uint64 = data[i].Timestamp/aggr*aggr
				if t != prevt && count > 0 {
					ans.X = append(ans.X, prevt)
					ans.Y = append(ans.Y, sum/float64(count))
					sum = 0
					count = 0
				}
				if value, ok := data[i].Value.(float64); ok {
					sum += value
					count++
				}
				prevt = t
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	enc := json.NewEncoder(w)
	err = enc.Encode(ans)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", err)
		log.Printf("Internal Server Error: %v", err)
		return
	}
}
