package main

import (
	"fmt"
	"os"
	"runtime"
	"flag"
	"log"
	"code.google.com/p/gcfg"
)

import "hammy"

// Debug and statistics
import (
	"net/http"
	_ "net/http/pprof"
	_ "expvar"
)

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

	if cfg.Log.HammyDFile != "" {
		logf, err := os.OpenFile(cfg.Log.HammyDFile,
			os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0666)

		if err != nil {
			log.Fatalf("Failed to open logfile: %v", err)
		}

		log.SetOutput(logf)
	}

	log.Printf("Initializing...")

	go func() {
		log.Println(http.ListenAndServe(cfg.Debug.HammyDAddr, nil))
	}()

	tg, err := hammy.NewCouchbaseTriggersGetter(cfg)
	if err != nil {
		log.Fatalf("CouchbaseTriggersGetter: %v", err)
	}

	sk, err := hammy.NewCouchbaseStateKeeper(cfg)
	if err != nil {
		log.Fatalf("CouchbaseStateKeeper: %v", err)
	}

	e := hammy.NewSPExecuter(cfg)

	cbp := hammy.CmdBufferProcessorImpl{}

	rh := hammy.RequestHandlerImpl{
		TGetter: tg,
		SKeeper: sk,
		Exec: e,
		CBProcessor: &cbp,
	}

	sb := hammy.NewSendBufferImpl(&rh, cfg)
	go sb.Listen()
	cbp.SBuffer = sb

	saver, err := hammy.NewCouchbaseSaver(cfg)
	if err != nil {
		log.Fatal("CouchbaseSaver: %v", err)
	}
	cbp.Saver = saver

	log.Printf("done.")
	log.Printf("Starting HTTP interface...")
	err = hammy.StartHttp(&rh, cfg)
	log.Fatal(err)
}
