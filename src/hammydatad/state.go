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

	var ans hammy.StateKeeperAnswer

	if obj == "__test" {
		ans = GenTestState()
	} else {
		ans = h.SKeeper.Get(obj)
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