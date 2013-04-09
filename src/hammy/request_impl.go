package hammy

import (
	"math/rand"
	"time"
	"fmt"
	"log"
	"reflect"
)

// Main data processor implementation
// You must call InitMetrics before use new object
type RequestHandlerImpl struct {
	// Interface for triggers retriving
	TGetter TriggersGetter
	// Interface for state storage
	SKeeper StateKeeper
	// Interface for executer
	Exec Executer
	// Interface for command commiter
	CBProcessor CmdBufferProcessor

	// Metrics
	ms *MetricSet
	mHandle *TimerMetric
	mCollisions *CounterMetric
}

// Initializes statistical metrics
func (rh *RequestHandlerImpl) InitMetrics(metricsNamespace string) {
	rh.ms = NewMetricSet(metricsNamespace, 30*time.Second)
	rh.mHandle = rh.ms.NewTimer("handle")
	rh.mCollisions = rh.ms.NewCounter("collisions")

	rand.Seed( time.Now().UTC().UnixNano())
}

// Internal struct for processing result
type hostProcessResult struct {
	Key string
	Err error
}

func (rh *RequestHandlerImpl) Handle(data IncomingData) (errs map[string]error) {
	// Allocate return value
	errs = make(map[string]error)

	// Is it possible?... May be...
	if (len(data) == 0) {
		return
	}

	// Statistics
	τ := rh.mHandle.NewObservation()
	defer func() { τ.End() } ()

	// Step 1: Loading triggers
	keys := make([]string, len(data))
	i := 0
	for k, _ := range data {
		keys[i] = k
		i++
	}

	triggers, err := rh.TGetter.MGet(keys)
	if err != nil {
		for _, k := range keys {
			errs[k] = fmt.Errorf("Failed to load trigger: %v", err)
		}
		return
	}

	if len(triggers) == 0 {
		return
	}

	// fix keys array
	keys = make([]string, len(triggers))
	i = 0
	for k := range triggers {
		keys[i] = k
		i++
	}

	// Step 2: Loading state
	states := rh.SKeeper.MGet(keys)
	if len(states) != len(triggers) {
		panic(fmt.Sprintf("Invalid length of StateKeeper answer (%v states for %v triggers)",
				len(states), len(triggers)))
	}

	// Step 3: Start trigger processing
	c := make(chan hostProcessResult) // result chanel
	runningTasks := 0
	for key, tr := range triggers {
		state := states[key]
		if state.Err != nil {
			errs[key] = fmt.Errorf("Failed to load state: %v", state.Err)
		} else {
			go rh.processTrigger(key, tr,
				state.State, state.Cas, data[key], c)
			runningTasks++
		}
	}

	// Step 4: Waiting until all triggers done
	if runningTasks > 0 {
		for res := range c {
			if res.Err != nil {
				errs[res.Key] = res.Err
			}
			runningTasks--
			if runningTasks == 0 {
				break
			}
		}
	}

	return
}

func (rh *RequestHandlerImpl) logErr(key string, message string) {
	cmdb := make(CmdBuffer, 1)
	cmdb[0].Cmd = "log"
	cmdb[0].Options = make(map[string]interface{})
	cmdb[0].Options["message"] = message
	err := rh.CBProcessor.Process(key, &cmdb)
	if err != nil {
		log.Printf("!! Error in cmd_buffer: %v", err)
	}
}

func (rh *RequestHandlerImpl) processTrigger(
		key string,
		trigger string,
		state State,
		cas *uint64,
		data IncomingHostData,
		c chan hostProcessResult) {
//
	ret := hostProcessResult{Key: key, Err: nil}
	defer func() {
		c <- ret
	}()
//
	newState, cmdb, err := rh.Exec.ProcessTrigger(key, trigger, &state, data)
	if err != nil {
		ret.Err = fmt.Errorf("Trigger processor error: %v", err)
		rh.logErr(key, err.Error())
		return
	}

	// Save state if it changes
	if !StatesEq(newState, &state) {
		retry, err := rh.SKeeper.Set(key, *newState, cas)
		if err != nil {
			ret.Err = fmt.Errorf("StateKeeper set operation failed: %v", err)
			return
		}
		if retry {
			// Ooops! Collision!
			// Starting loop of retries
			for {
				// Statistics
				rh.mCollisions.Add(1)

				// random sleep
				t := rand.Intn(1000)
				time.Sleep(time.Duration(t) * time.Millisecond)

				// Retry...
				ans := rh.SKeeper.Get(key)
				if ans.Err != nil {
					ret.Err = fmt.Errorf("StateKeeper get operation failed: %v", ans.Err)
					return
				}
				state = ans.State
				cas = ans.Cas

				newState, cmdb, err = rh.Exec.ProcessTrigger(key, trigger, &state, data)
				if err != nil {
					ret.Err = fmt.Errorf("Trigger processor error: %v", err)
					rh.logErr(key, err.Error())
					return
				}

				// Save state if it changes
				if !StatesEq(newState, &state) {
					retry, err = rh.SKeeper.Set(key, *newState, cas)
					if err != nil {
						ret.Err = fmt.Errorf("StateKeeper set operation failed: %v", err)
						return
					}
					if !retry {
						// Work is done! No more need to retry!
						break
					}
				} else {
					// %)
					break
				}
			}
		}
	}

	// Commit cmdbuffer
	err = rh.CBProcessor.Process(key, cmdb)
	if err != nil {
		ret.Err = fmt.Errorf("Failed to process cmdbuffer: %v", err)
		return
	}

	return
}

// Compare two States
// true if equal, false otherwise
func StatesEq(a *State, b *State) bool {
	return reflect.DeepEqual(a, b)
}
