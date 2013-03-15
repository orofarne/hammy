package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"encoding/json"
)

import _ "hammy"

type Answer struct {
	X []uint64
	Y []float64
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

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

	obj_a := q["object"]
	if len(obj_a) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	obj := obj_a[0]
	if obj == "" {
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

	if obj != "test" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	} else {
		ans := new(Answer)
		switch key {
			case "sin":
				n := to - from + 1
				ans.X = make([]uint64, n)
				ans.Y = make([]float64, n)
				var i uint64
				for i = 0; i < n; i++ {
					ans.X[i] = from + i
				}
				for i = 0; i < n; i++ {
					ans.Y[i] = math.Sin(float64(from + i) / 100.0 * math.Pi)
				}
			default:
				http.Error(w, "Not Found", http.StatusNotFound)
				return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		err := enc.Encode(ans)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Fprintf(w, "%v\n", err);
			log.Printf("Internal Server Error: %v", err)
		}
	}
}

func main() {
	http.HandleFunc("/data", reqHandler)

	log.Fatal(http.ListenAndServe("localhost:8093", nil))
}
