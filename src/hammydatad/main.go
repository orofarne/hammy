package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"math"
	"time"
	"runtime"
	"net/http"
	"encoding/json"
	"code.google.com/p/gcfg"
)

import "hammy"

//Debug and statistics
import (
	_ "net/http/pprof"
	_ "expvar"
)

type Answer struct {
	X []uint64
	Y []float64
}

//Returns data for tests
type TestDataReader struct {
}

func (tr *TestDataReader) Read(objKey string, itemKey string, from uint64, to uint64) (data []hammy.IncomingValueData, err error) {
	if objKey != "__test" {
		panic(fmt.Sprintf("Requested test data for key %#v", objKey))
	}

	switch itemKey {
		case "sin":
			n := to - from + 1
			data := make([]hammy.IncomingValueData, n)
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

//Http server object
type HttpServer struct{
	//Data reader
	DReader hammy.DataReader
}

//Request handler
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
	ans.Y = make([]float64, n)
	for i := 0; i < n; i++ {
		ans.X[i] = data[i].Timestamp
		var converted bool
		ans.Y[i], converted = data[i].Value.(float64)
		if !converted {
			err := fmt.Errorf("Failed to convert `%#v` to float64", data[i].Value)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Fprintf(w, "%v\n", err)
			log.Printf("Internal Server Error: %v", err)
			return
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

//Parse comand-line and fill config
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

	dr, err := hammy.NewCouchbaseDataReader(cfg)
	if err != nil {
		log.Fatalf("NewCouchbaseDataReader: %v", err)
	}

	h := &HttpServer{
		DReader: dr,
	}

	//Setup server
	s := &http.Server{
		Addr:				cfg.DataHttp.Addr,
		Handler:			h,
		ReadTimeout:		30 * time.Second,
		WriteTimeout:		30 * time.Second,
		MaxHeaderBytes:		1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
