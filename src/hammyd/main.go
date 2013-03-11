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

	log.Printf("Initializing...")

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

	cbp.RHandler = &rh

	log.Printf("Starting HTTP interface...")
	err = hammy.StartHttp(&rh, cfg)
	log.Fatal(err)
}
