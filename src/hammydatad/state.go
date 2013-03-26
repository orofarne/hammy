package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
)

import "hammy"

func (h *HttpServer) ServeState(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

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

	var ans hammy.StateKeeperAnswer

	if host == "__test" {
		ans = GenTestState()
	} else {
		ans = h.SKeeper.Get(host)
		if ans.Err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Fprintf(w, "%v\n", ans.Err)
			log.Printf("Internal Server Error: %v", ans.Err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	enc := json.NewEncoder(w)
	err := enc.Encode(ans.State)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", err)
		log.Printf("Internal Server Error: %v", err)
		return
	}
}