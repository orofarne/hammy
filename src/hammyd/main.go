package main

import (
	"fmt"
	"os"
	"io"
	"bufio"
	"runtime"
	"flag"
	"log"
	"code.google.com/p/gcfg"
)

import "hammy"

// Debug and statistics
import (
	"net/http"
//	_ "net/http/pprof"
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

	var cmdOutF io.Writer
	if cfg.Log.CmdOutputFile != "" {
		f, err := os.OpenFile(cfg.Log.CmdOutputFile,
			os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0666)

		if err != nil {
			log.Fatalf("Failed to open cmdoutput file: %v", err)
		}

		cmdOutF = bufio.NewWriter(f)
	} else {
		cmdOutF = bufio.NewWriter(os.Stdout)
	}

	go func() {
		log.Println(http.ListenAndServe(cfg.Debug.HammyDAddr, nil))
	}()

	var tg hammy.TriggersGetter
	if cfg.CouchbaseTriggers.Active {
		tg, err = hammy.NewCouchbaseTriggersGetter(cfg)
		if err != nil {
			log.Fatalf("CouchbaseTriggersGetter: %v", err)
		}
	}
	if cfg.MySQLTriggers.Active {
		tg, err = hammy.NewMySQLTriggersGetter(cfg)
		if err != nil {
			log.Fatal("MySQLTriggersGetter: %v", err)
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

	e := hammy.NewSPExecuter(cfg, "spexecuter")

	cbp := hammy.CmdBufferProcessorImpl{
		CmdOutput: cmdOutF,
	}

	rh := hammy.RequestHandlerImpl{
		TGetter: tg,
		SKeeper: sk,
		Exec: e,
		CBProcessor: &cbp,
	}
	rh.InitMetrics("request_handler")

	sb := hammy.NewSendBufferImpl(&rh, cfg, "send_buffer")
	go sb.Listen()
	cbp.SBuffer = sb

	var saver hammy.SendBuffer
	if cfg.CouchbaseSaver.Active {
		saver, err = hammy.NewCouchbaseSaver(cfg)
		if err != nil {
			log.Fatal("CouchbaseSaver: %v", err)
		}
	}
	if cfg.MySQLSaver.Active {
		saver, err = hammy.NewMySQLSaver(cfg)
		if err != nil {
			log.Fatal("MySQLSaver: %v", err)
		}
	}

	cbp.Saver = saver

	log.Printf("done.")
	log.Printf("Starting HTTP interface...")
	err = hammy.StartHttp(&rh, cfg, "incoming_http")
	log.Fatal(err)
}
