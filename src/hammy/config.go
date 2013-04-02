package hammy

import "fmt"

// Programm configuration
type Config struct {
	// Http interface for incoming data
	IncomingHttp struct {
		// Addr for incomming-data
		// e.g. "0.0.0.0:4000" or ":4000" for ipv6
		Addr string
	}
	// Http interface for data requests
	DataHttp struct {
		// Addr for data request
		// e.g. "0.0.0.0:4000" or ":4000" for ipv6
		Addr string
		// Http prefix
		// e.g. /data (default)
		Prefix string
	}
	// Logging options
	Log struct {
		// Files for logging (stderr if empty)
		// For hammyd daemon
		HammyDFile string
		// For hammydatad daemon
		HammyDataDFile string
	}
	// Debug and statistics
	Debug struct {
		// Addrs for debug and statistic information
		// e.g. "localhost:6060" (default)
		// For hammyd daemon
		HammyDAddr string
		// For hammydatad daemon
		HammyDataDAddr string
	}
	// Workers
	Workers struct {
		// Count of workers
		PoolSize uint
		// Worker cmd
		CmdLine string
		// Worker live limit
		MaxIter uint
		// Worker timeout (before kill)
		Timeout uint
	}
	// Send buffer
	SendBuffer struct {
		SleepTime float32
	}
	// Coucbase for triggers configuration
	CouchbaseTriggers struct {
		// Use this implementation
		Active bool
		// e.g. "http://dev-couchbase.example.com:8091/"
		ConnectTo string
		// e.g. "default"
		Pool string
		// e.g. "default"
		Bucket string
	}
	// Coucbase for state storage
	CouchbaseStates struct {
		// Use this implementation
		Active bool
		// e.g. "http://dev-couchbase.example.com:8091/"
		ConnectTo string
		// e.g. "default"
		Pool string
		// e.g. "default"
		Bucket string
		// TTL in seconds, default 86400 (day)
		Ttl int
	}
	// Data saver
	CouchbaseSaver struct {
		// Use this implementation
		Active bool
		// e.g. "http://dev-couchbase.example.com:8091/"
		ConnectTo string
		// e.g. "default"
		Pool string
		// e.g. "default"
		Bucket string
		// Internal write queue size
		QueueSize uint
		// Connections for saving
		SavePoolSize uint
		// TTL in seconds, default 259200 (3 days)
		Ttl int
	}
	// Data reader
	CouchbaseDataReader struct {
		// Use this implementation
		Active bool
		// e.g. "http://dev-couchbase.example.com:8091/"
		ConnectTo string
		// e.g. "default"
		Pool string
		// e.g. "default"
		Bucket string
	}
	// MySQL for triggers configuration
	MySQLTriggers struct {
		// Use this implementation
		Active bool
		// Database to connect
		Database string
		// DB user
		User string
		// DB user password
		Password string
		// table that contains triggers (host, trigger)
		Table string
		// Limit for parallel connections
		MaxConn int
	}
	// MySQL for states
	MySQLStates struct {
		// Use this implementation
		Active bool
		// Database to connect
		Database string
		// DB user
		User string
		// DB user password
		Password string
		// table that contains states (host, state, cas)
		Table string
		// Limit for parallel connections
		MaxConn int
	}
	// MySQL historical data saver
	MySQLSaver struct {
		// Use this implementation
		Active bool
		// Database to connect
		Database string
		// DB user
		User string
		// DB user password
		Password string
		// table that contains numeric history
		Table string
		// table that contains text history
		LogTable string
		// table that contains hosts
		HostTable string
		// table that contains items
		ItemTable string
		// Limit for parallel connections
		MaxConn int
	}
	// MySQL historical data reader
	MySQLDataReader struct {
		// Use this implementation
		Active bool
		// Database to connect
		Database string
		// DB user
		User string
		// DB user password
		Password string
		// table that contains numeric history
		Table string
		// table that contains text history
		LogTable string
		// table that contains hosts
		HostTable string
		// table that contains items
		ItemTable string
		// Limit for parallel connections
		MaxConn int
	}
}

// Setup defaults for empty values in configs
// Returns an error if mandatory field omited
func SetConfigDefaults(cfg *Config) error {
	// Section [IncomingHttp]
	if cfg.IncomingHttp.Addr == "" { cfg.IncomingHttp.Addr = ":4000" }

	// Section [DataHttp]
	if cfg.DataHttp.Addr == "" { cfg.DataHttp.Addr = ":4001" }
	if cfg.DataHttp.Prefix == "" { cfg.DataHttp.Prefix = "/data" }

	// Section [Log]

	// Section [Debug]
	if cfg.Debug.HammyDAddr == "" { cfg.Debug.HammyDAddr = "localhost:6060" }
	if cfg.Debug.HammyDataDAddr == "" { cfg.Debug.HammyDataDAddr = "localhost:6061" }

	// Section [SendBuffer]
	if cfg.SendBuffer.SleepTime == 0.0 { cfg.SendBuffer.SleepTime = 10.0 }

	// Section [Workers]
	if cfg.Workers.PoolSize == 0 { cfg.Workers.PoolSize = 5 }
	if cfg.Workers.CmdLine == "" { return fmt.Errorf("Empty cfg.Workers.CmdLine") }
	if cfg.Workers.MaxIter == 0 { cfg.Workers.MaxIter = 1000 }
	if cfg.Workers.Timeout == 0 { cfg.Workers.Timeout = 1 }

	// Section [CouchbaseTriggers]
	if cfg.CouchbaseTriggers.Active {
		if cfg.CouchbaseTriggers.ConnectTo == "" { return fmt.Errorf("Empty cfg.CouchbaseTriggers.ConnectTo") }
		if cfg.CouchbaseTriggers.Pool == "" { cfg.CouchbaseTriggers.Pool = "default" }
		if cfg.CouchbaseTriggers.Bucket == "" { return fmt.Errorf("Empty cfg.CouchbaseTriggers.Bucket") }
	}

	// Section [CouchbaseStates]
	if cfg.CouchbaseStates.Active {
		if cfg.CouchbaseStates.ConnectTo == "" { return fmt.Errorf("Empty cfg.CouchbaseStates.ConnectTo") }
		if cfg.CouchbaseStates.Pool == "" { cfg.CouchbaseStates.Pool = "default" }
		if cfg.CouchbaseStates.Bucket == "" { return fmt.Errorf("Empty cfg.CouchbaseStates.Bucket") }
		if cfg.CouchbaseStates.Ttl == 0 { cfg.CouchbaseStates.Ttl = 86400 }
	}

	// Section [CouchbaseSaver]
	if cfg.CouchbaseSaver.Active {
		if cfg.CouchbaseSaver.ConnectTo == "" { return fmt.Errorf("Empty cfg.CouchbaseSaver.ConnectTo") }
		if cfg.CouchbaseSaver.Pool == "" { cfg.CouchbaseSaver.Pool = "default" }
		if cfg.CouchbaseSaver.Bucket == "" { return fmt.Errorf("Empty cfg.CouchbaseSaver.Bucket") }
		if cfg.CouchbaseSaver.QueueSize == 0 { cfg.CouchbaseSaver.QueueSize = 10000 }
		if cfg.CouchbaseSaver.SavePoolSize == 0 { cfg.CouchbaseSaver.SavePoolSize = 10 }
		if cfg.CouchbaseSaver.Ttl == 0 { cfg.CouchbaseSaver.Ttl = 259200 }
	}

	// Section [CouchbaseDataReader]
	if cfg.CouchbaseDataReader.Active {
		if cfg.CouchbaseDataReader.ConnectTo == "" { return fmt.Errorf("Empty cfg.CouchbaseDataReader.ConnectTo") }
		if cfg.CouchbaseDataReader.Pool == "" { cfg.CouchbaseDataReader.Pool = "default" }
		if cfg.CouchbaseDataReader.Bucket == "" { return fmt.Errorf("Empty cfg.CouchbaseDataReader.Bucket") }
	}

	// Section [MySQLTriggers]
	if cfg.MySQLTriggers.Active {
		if cfg.MySQLTriggers.Database == "" { return fmt.Errorf("Empty cfg.MySQLTriggers.Database") }
		if cfg.MySQLTriggers.User == "" { return fmt.Errorf("Empty cfg.MySQLTriggers.User") }
		if cfg.MySQLTriggers.Table == "" { return fmt.Errorf("Empty cfg.MySQLTriggers.Table") }
		if cfg.MySQLTriggers.MaxConn == 0 { cfg.MySQLTriggers.MaxConn = 10 }
	}

	// Section [MySQLStates]
	if cfg.MySQLStates.Active {
		if cfg.MySQLStates.Database == "" { return fmt.Errorf("Empty cfg.MySQLStates.Database") }
		if cfg.MySQLStates.User == "" { return fmt.Errorf("Empty cfg.MySQLStates.User") }
		if cfg.MySQLStates.Table == "" { return fmt.Errorf("Empty cfg.MySQLStates.Table") }
		if cfg.MySQLStates.MaxConn == 0 { cfg.MySQLStates.MaxConn = 10 }
	}

	// Section [MySQLSaver]
	if cfg.MySQLSaver.Active {
		if cfg.MySQLSaver.Database == "" { return fmt.Errorf("Empty cfg.MySQLSaver.Database") }
		if cfg.MySQLSaver.User == "" { return fmt.Errorf("Empty cfg.MySQLSaver.User") }
		if cfg.MySQLSaver.Table == "" { return fmt.Errorf("Empty cfg.MySQLSaver.Table") }
		if cfg.MySQLSaver.LogTable == "" { return fmt.Errorf("Empty cfg.MySQLSaver.LogTable") }
		if cfg.MySQLSaver.HostTable == "" { return fmt.Errorf("Empty cfg.MySQLSaver.HostTable") }
		if cfg.MySQLSaver.ItemTable == "" { return fmt.Errorf("Empty cfg.MySQLSaver.ItemTable") }
		if cfg.MySQLSaver.MaxConn == 0 { cfg.MySQLSaver.MaxConn = 10 }
	}

	// Section [MySQLDataReader]
	if cfg.MySQLDataReader.Active {
		if cfg.MySQLDataReader.Database == "" { return fmt.Errorf("Empty cfg.MySQLDataReader.Database") }
		if cfg.MySQLDataReader.User == "" { return fmt.Errorf("Empty cfg.MySQLDataReader.User") }
		if cfg.MySQLDataReader.Table == "" { return fmt.Errorf("Empty cfg.MySQLDataReader.Table") }
		if cfg.MySQLDataReader.LogTable == "" { return fmt.Errorf("Empty cfg.MySQLDataReader.LogTable") }
		if cfg.MySQLDataReader.HostTable == "" { return fmt.Errorf("Empty cfg.MySQLDataReader.HostTable") }
		if cfg.MySQLDataReader.ItemTable == "" { return fmt.Errorf("Empty cfg.MySQLDataReader.ItemTable") }
		if cfg.MySQLDataReader.MaxConn == 0 { cfg.MySQLDataReader.MaxConn = 10 }
	}

	// Counts
	// 1) TriggersGetter
	{
		k := 0

		if cfg.CouchbaseTriggers.Active { k++ }
		if cfg.MySQLTriggers.Active { k++ }

		if k != 1 {
			return fmt.Errorf("Invalid count of active TriggersGetter drivers: %d", k)
		}
	}
	// 2) StateKeeper
	{
		k := 0

		if cfg.CouchbaseStates.Active { k++ }
		if cfg.MySQLStates.Active { k++ }

		if k != 1 {
			return fmt.Errorf("Invalid count of active StateKeeper drivers: %d", k)
		}
	}
	// 3) DataSaver
	{
		k := 0

		if cfg.CouchbaseSaver.Active { k++ }
		if cfg.MySQLSaver.Active { k++ }

		if k != 1 {
			return fmt.Errorf("Invalid count of active DataSaver drivers: %d", k)
		}
	}
	// 4) DataReader
	{
		k := 0

		if cfg.CouchbaseDataReader.Active { k++ }
		if cfg.MySQLDataReader.Active { k++ }

		if k != 1 {
			return fmt.Errorf("Invalid count of active DataReader drivers: %d", k)
		}
	}

	return nil
}
