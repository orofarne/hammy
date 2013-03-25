package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"time"
	"runtime"
	"strings"
	"net/http"
	"code.google.com/p/gcfg"
)

import "hammy"

// Debug and statistics
import (
	_ "net/http/pprof"
	_ "expvar"
)

// Http server object
type HttpServer struct{
	// Prefix
	Prefix string
	// Data reader
	DReader hammy.DataReader
	// State keeper
	SKeeper hammy.StateKeeper
}

// Request router
func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.Prefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	path := r.URL.Path[len(h.Prefix):]

	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	switch path {
		case "/values":
			h.ServeValues(w, r) // defined in values.go
			return
		case "/state":
			h.ServeState(w, r) // defined in state.go
			return
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
			return
	}
	panic("?!")
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

	var sk hammy.StateKeeper
	if cfg.CouchbaseStates.Active {
		sk, err = hammy.NewCouchbaseStateKeeper(cfg)
		if err != nil {
			log.Fatalf("CouchbaseStateKeeper: %v", err)
		}
	}
	if cfg.MySQLStates.Active {
		sk, err = hammy.NewMySQLStateKeeper(cfg)
		if err != nil {
			log.Fatalf("MySQLStateKeeper: %v", err)
		}
	}

	h := &HttpServer{
		Prefix: cfg.DataHttp.Prefix,
		DReader: dr,
		SKeeper: sk,
	}

	log.Printf("done.")
	log.Printf("Starting HTTP interface...")

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
