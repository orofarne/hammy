package hammy

import "fmt"

//Programm configuration
type Config struct {
	//Http interface section
	Http struct {
		//Addr for incomming-data
		//e.g. "0.0.0.0:4000" or ":4000" for ipv6
		Addr string
	}
	//Logging options
	Log struct {
		//File for logging (stderr if empty)
		File string
	}
	//Debug and statistics
	Debug struct {
		//Addr for debug and statistic information
		//e.g. "localhost:6060" (default)
		Addr string
	}
	//Workers
	Workers struct {
		//Count of workers
		PoolSize uint
		//Worker cmd
		CmdLine string
		//Worker live limit
		MaxIter uint
	}
	//Send buffer
	SendBuffer struct {
		SleepTime float32
	}
	//Coucbase for triggers configuration
	CouchbaseTriggers struct {
		//e.g. "http://dev-couchbase.example.com:8091/"
		ConnectTo string
		//e.g. "default"
		Pool string
		//e.g. "default"
		Bucket string
	}
	//Coucbase for state storage
	CouchbaseStates struct {
		//e.g. "http://dev-couchbase.example.com:8091/"
		ConnectTo string
		//e.g. "default"
		Pool string
		//e.g. "default"
		Bucket string
		//TTL in seconds, default 86400
		Ttl int
	}
	//Data saver
	CouchbaseSaver struct {
		//e.g. "http://dev-couchbase.example.com:8091/"
		ConnectTo string
		//e.g. "default"
		Pool string
		//e.g. "default"
		Bucket string
		//Internal write queue size
		QueueSize uint
		//Connections for saving
		SavePoolSize uint
	}
}

//Setup defaults for empty values in configs
//Returns an error if mandatory field omited
func SetConfigDefaults(cfg *Config) error {
	//Section [Http]
	if cfg.Http.Addr == "" { cfg.Http.Addr = ":4000" }

	//Section [Log]

	//Section [Debug]
	if cfg.Debug.Addr == "" { cfg.Debug.Addr = "localhost:6060" }

	//Section [SendBuffer]
	if cfg.SendBuffer.SleepTime == 0.0 { cfg.SendBuffer.SleepTime = 10.0 }

	//Section [Workers]
	if cfg.Workers.PoolSize == 0 { cfg.Workers.PoolSize = 5 }
	if cfg.Workers.CmdLine == "" { return fmt.Errorf("Empty cfg.Workers.CmdLine") }
	if cfg.Workers.MaxIter == 0 { cfg.Workers.MaxIter = 1000 }

	//Section [CouchbaseTriggers]
	if cfg.CouchbaseTriggers.ConnectTo == "" { return fmt.Errorf("Empty cfg.CouchbaseTriggers.ConnectTo") }
	if cfg.CouchbaseTriggers.Pool == "" { cfg.CouchbaseTriggers.Pool = "default" }
	if cfg.CouchbaseTriggers.Bucket == "" { return fmt.Errorf("Empty cfg.CouchbaseTriggers.Bucket") }

	//Section [CouchbaseStates]
	if cfg.CouchbaseStates.ConnectTo == "" { return fmt.Errorf("Empty cfg.CouchbaseStates.ConnectTo") }
	if cfg.CouchbaseStates.Pool == "" { cfg.CouchbaseStates.Pool = "default" }
	if cfg.CouchbaseStates.Bucket == "" { return fmt.Errorf("Empty cfg.CouchbaseStates.Bucket") }
	if cfg.CouchbaseStates.Ttl == 0 { cfg.CouchbaseStates.Ttl = 86400 }

	//Section [CouchbaseSaver]
	if cfg.CouchbaseSaver.ConnectTo == "" { return fmt.Errorf("Empty cfg.CouchbaseSaver.ConnectTo") }
	if cfg.CouchbaseSaver.Pool == "" { cfg.CouchbaseSaver.Pool = "default" }
	if cfg.CouchbaseSaver.Bucket == "" { return fmt.Errorf("Empty cfg.CouchbaseSaver.Bucket") }
	if cfg.CouchbaseSaver.QueueSize == 0 { cfg.CouchbaseSaver.QueueSize = 10000 }
	if cfg.CouchbaseSaver.SavePoolSize == 0 { cfg.CouchbaseSaver.SavePoolSize = 10 }

	return nil
}
