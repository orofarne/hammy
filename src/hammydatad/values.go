package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
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
	ans.X = make([]uint64, n)
	ans.Y = make([]interface{}, n)
	for i := 0; i < n; i++ {
		ans.X[i] = data[i].Timestamp
		ans.Y[i] = data[i].Value
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