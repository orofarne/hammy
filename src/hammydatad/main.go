package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"time"
	"runtime"
	"net/http"
	"encoding/json"
	"code.google.com/p/gcfg"
)

import "hammy"

// Debug and statistics
import (
	_ "net/http/pprof"
	_ "expvar"
)

type Answer struct {
	X []uint64
	Y []interface{}
}

// Http server object
type HttpServer struct{
	// Data reader
	DReader hammy.DataReader
}

// Request handler
func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	var dataReader hammy.DataReader
	if obj == "__test" {
		dataReader = &TestDataReader{}
	} else {
		dataReader = h.DReader
	}

	data, err := dataReader.Read(obj, key, from, to)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", err)
		log.Printf("Internal Server Error: %v", err)
		return
	}

	ans := new(Answer)
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

// Parse comand-line and fill config
func loadConfig(cfg *hammy.Config) {
	var configFile string

	flag.StringVar(&configFile, "config", "", "Config file path")
	flag.StringVar(&configFile, "c", "", "Config file path (short)")
	flag.Parse()

	if configFile == "" {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := gcfg.ReadFileInto(cfg, configFile); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var cfg hammy.Config

	loadConfig(&cfg)
	err := hammy.SetConfigDefaults(&cfg)
	if err != nil {
		log.Fatalf("Inavalid config: %v", err)
	}

	if cfg.Log.HammyDataDFile != "" {
		logf, err := os.OpenFile(cfg.Log.HammyDataDFile,
			os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0666)

		if err != nil {
			log.Fatalf("Failed to open logfile: %v", err)
		}

		log.SetOutput(logf)
	}

	log.Printf("Initializing...")

	go func() {
		log.Println(http.ListenAndServe(cfg.Debug.HammyDataDAddr, nil))
	}()

	var dr hammy.DataReader
	if cfg.CouchbaseDataReader.Active {
		dr, err = hammy.NewCouchbaseDataReader(cfg)
		if err != nil {
			log.Fatalf("NewCouchbaseDataReader: %v", err)
		}
	}
	if cfg.MySQLDataReader.Active {
		dr, err = hammy.NewMySQLDataReader(cfg)
		if err != nil {
			log.Fatalf("NewMySQLDataReader: %v", err)
		}
	}

	h := &HttpServer{
		DReader: dr,
	}

	// Setup server
	s := &http.Server{
		Addr:				cfg.DataHttp.Addr,
		Handler:			h,
		ReadTimeout:		30 * time.Second,
		WriteTimeout:		30 * time.Second,
		MaxHeaderBytes:		1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
