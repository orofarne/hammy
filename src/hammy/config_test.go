package hammy

import "testing"

func TestSetConfigDefaults(t *testing.T) {
	var cfg Config
	err := SetConfigDefaults(&cfg)
	if err == nil {
		t.Errorf("Error should not be nil")
	}
	//Mandatory fields:
	cfg.Workers.CmdLine = "/bin/ls"
	cfg.CouchbaseTriggers.ConnectTo = "http://localhost:8091/"
	cfg.CouchbaseTriggers.Bucket = "default"
	cfg.CouchbaseStates.ConnectTo = "http://localhost:8091/"
	cfg.CouchbaseStates.Bucket = "default"
	cfg.CouchbaseSaver.ConnectTo = "http://localhost:8091/"
	cfg.CouchbaseSaver.Bucket = "default"
	cfg.CouchbaseDataReader.ConnectTo = "http://localhost:8091/"
	cfg.CouchbaseDataReader.Bucket = "default"

	//Retry...
	err = SetConfigDefaults(&cfg)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if cfg.IncomingHttp.Addr != ":4000" {
		t.Errorf("cfg.ImcomingHttp.Addr = %#v, expected %#v", cfg.IncomingHttp.Addr, ":4000")
	}
	if cfg.DataHttp.Addr != ":4001" {
		t.Errorf("cfg.DataHttp.Addr = %#v, expected %#v", cfg.DataHttp.Addr, ":4001")
	}
	if cfg.CouchbaseStates.Ttl != 86400 {
		t.Errorf("cfg.CouchbaseStates.Ttl = %#v, expected 86400", cfg.CouchbaseStates.Ttl)
	}
	if cfg.SendBuffer.SleepTime != 10.0 {
		t.Errorf("cfg.SendBuffer.SleepTime = %#v, expected %#v", cfg.SendBuffer.SleepTime, 10.0)
	}
}
