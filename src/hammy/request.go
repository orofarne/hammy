package hammy

type IncomingValueData struct {
	Timestamp uint64
	Value interface{}
}

type IncomingHostData map[string][]IncomingValueData

// Type for incoming monitoring data
// Format (in json notation):
//  {
//    "host1": {
//      "key1.1": [{
//        "Timestamp": 1361785778,
//        "Value": 3.14
//      }]
//    },
//    "host2": {
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
type IncomingData map[string]IncomingHostData

// Type for incoming monitoring data request
type IncomingMessage struct {
	// Incoming monitoring data
	Data IncomingData
	// Processing level (0 for new data)
	// Increments after each resend
	Level uint32
}

// Response
type ResponseMessage struct {
	Errors map[string]string
}

// Interface for incoming data handler
type RequestHandler interface {
	Handle(data IncomingData) map[string]error
}

// Type for state of an host
// maps keys to values with timestamps of last update
// Format (in json notation):
//  {
//    "key1": {
//      "Value": 10.3,
//      "LastUpdate": 1361785927
//    },
//    "key2": {
//      "Value": "booo!",
//      "LastUpdate": 1361785778
//    }
type State map[string]StateElem

// Type for State element
type StateElem struct {
	LastUpdate uint64
	Value interface{}
}

func NewState() *State {
	s := make(State)
	return &s
}
