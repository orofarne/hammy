package hammy

// Interface for trigger configuration
type TriggersGetter interface {
	MGet(keys []string) (triggers map[string]string, err error)
}

// Answer of StateKeeper's get requests
type StateKeeperAnswer struct {
	State
	Cas *uint64
	Err error
}

// Interface for state storage
type StateKeeper interface {
	Get(key string) StateKeeperAnswer
	MGet(keys []string) map[string]StateKeeperAnswer
	Set(key string, data State, cas *uint64) (retry bool, err error)
}

// Interface for internal sendbuffers and for external data savers
type SendBuffer interface {
	Push(data *IncomingData)
}

// Reads data from write cache or storage
type DataReader interface {
	Read(hostKey string, itemKey string, from uint64, to uint64) (data []IncomingValueData, err error)
}