package hammy

type IncomingValueData struct {
	Timestamp uint64
	Value interface{}
}

type IncomingObjectData map[string][]IncomingValueData

//Type for incoming monitoring data
//Format (in json notation):
//  {
//    "object1": {
//      "key1.1": [{
//        "Timestamp": 1361785778,
//        "Value": 3.14
//      }]
//    },
//    "object2": {
//      "key2.1": [{
//        "Timestamp": 1361785817,
//        "Value": "test string"
//      }],
//      "key2.2": [{
//        "Timestamp": 1361785858,
//        "Value": 12345
//      },
//      {
//        "Timestamp": 1361785927,
//        "Value": 999.3
//      }]
//    }
//  }
type IncomingData map[string]IncomingObjectData

func NewIncomingData() IncomingData {
	return make(map[string]IncomingObjectData)
}

//Interface for incoming data handler
type RequestHandler interface {
	Handle(data IncomingData) map[string]error
}

//Interface for trigger configuration
type TriggersGetter interface {
	MGet(keys []string) (triggers map[string]string, err error)
}

//Type for state of an object
//maps keys to values with timestamps of last update
//Format (in json notation):
//  {
//    "key1": {
//      "Value": 10.3,
//      "LastUpdate": 1361785927
//    },
//    "key2": {
//      "Value": "booo!",
//      "LastUpdate": 1361785778
//    }
type State map[string]struct {
	LastUpdate uint64
	Value interface{}
}

func NewState() State {
	return make(map[string]struct {
		LastUpdate uint64
		Value interface{}
	})
}

//Answer of StateKeeper's get requests
type StateKeeperAnswer struct {
	State
	Cas *uint64
	Err error
}

//Interface for state storage
type StateKeeper interface {
	Get(key string) StateKeeperAnswer
	MGet(keys []string) map[string]StateKeeperAnswer
	Set(key string, data State, cas *uint64) (retry bool, err error)
}
