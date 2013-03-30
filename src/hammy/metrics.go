package hammy

import (
	"expvar"
	"time"
	"sync"
	"encoding/json"
)

type MetricType int

const (
	METRIC_COUNTER MetricType = iota
	METRIC_TIMER
)

type metric struct {
	Type MetricType
	Name string
	Value interface{}
}

type metricState struct {
	Type MetricType
	Value interface{}
}

type metricTimerState struct {
	Counter uint64
	Sum uint64
}

// Namespaced set of metrics
type MetricSet struct {
	results map[string]interface{}
	states map[string]*metricState
	metricChan chan metric
	mutex sync.Mutex
	tickTime time.Duration
	prevTick time.Time
}

// Creates new metric set or panic if namespace exists
func NewMetricSet(namespace string, tickTime time.Duration) *MetricSet {
	ms := new(MetricSet)
	ms.results = make(map[string]interface{})
	ms.states = make(map[string]*metricState)
	ms.metricChan = make(chan metric, 100)
	ms.tickTime = tickTime

	var exportFunc expvar.Func
	exportFunc = func() interface{} {
		return ms.getVars()
	}
	expvar.Publish(namespace, exportFunc)

	go ms.collect()

	return ms
}

func (ms *MetricSet) getVars() interface{} {
	ms.mutex.Lock()
	defer func() { ms.mutex.Unlock() }()
	//FIXME achtung!
	var res interface{}
	buf, err := json.Marshal(ms.results)
	if err != nil { panic(err) }
	err = json.Unmarshal(buf, &res)
	if err != nil { panic(err) }
	return res
}

func (ms *MetricSet) collect() {
	ticker := time.Tick(ms.tickTime)
	ms.prevTick = time.Now()

	for {
		select {
			case m := <- ms.metricChan:
				ms.collectMetric(m)
			case t := <- ticker:
				Δ := t.Sub(ms.prevTick)
				ms.prevTick = t
				ms.doResults(Δ)
		}
	}
}

func (ms *MetricSet) collectMetric(m metric) {
	state, found := ms.states[m.Name]

	if !found {
		state = &metricState{
			Type: m.Type,
		}
	} else {
		if state.Type != m.Type {
			panic("Invalid metric type")
		}
	}

	switch m.Type {
		case METRIC_COUNTER:
			if state.Value == nil {
				state.Value = m.Value.(uint64)
			} else {
				state.Value = state.Value.(uint64) + m.Value.(uint64)
			}
		case METRIC_TIMER:
			τ := uint64(m.Value.(time.Duration).Nanoseconds())
			if state.Value == nil {
				state.Value = metricTimerState{
					Counter: 1,
					Sum: τ,
				}
			} else {
				pState := state.Value.(metricTimerState)
				state.Value = metricTimerState{
					Counter: (pState.Counter + 1),
					Sum: (pState.Sum + τ),
				}
			}
	}

	ms.states[m.Name] = state
}

func (ms *MetricSet) doResults(Δ time.Duration) {
	ms.mutex.Lock()
	defer func() { ms.mutex.Unlock() }()

	for k, v := range ms.states {
		switch v.Type {
			case METRIC_COUNTER:
				ms.results[k + "#rps"] = float64(v.Value.(uint64)) / Δ.Seconds()
				v.Value = uint64(0)
			case METRIC_TIMER:
				tState := v.Value.(metricTimerState)
				var τ float64
				if tState.Counter != 0 {
					τ = (float64(tState.Sum) / float64(tState.Counter)) / float64(1000000000)
				} else {
					τ = 0
				}
				ms.results[k + "#rps"] = float64(tState.Counter) / Δ.Seconds()
				ms.results[k + "_avgtime#s"] = τ
				v.Value = metricTimerState{}
		}
	}
}

type CounterMetric struct {
	name string
	c chan metric
}

func (ms *MetricSet) NewCounter(name string) *CounterMetric {
	m := new(CounterMetric)
	m.name = name
	m.c = ms.metricChan
	return m
}

func (m *CounterMetric)Add(n uint64) {
	m.c <- metric{
		Type: METRIC_COUNTER,
		Name: m.name,
		Value: n,
	}
}

type TimerMetric struct {
	name string
	c chan metric
}

func (ms *MetricSet) NewTimer(name string) *TimerMetric {
	m := new(TimerMetric)
	m.name = name
	m.c = ms.metricChan
	return m
}

func (m *TimerMetric) Add(τ time.Duration) {
	m.c <- metric{
		Type: METRIC_TIMER,
		Name: m.name,
		Value: τ,
	}
}

type TimerMetricObservation struct {
	m *TimerMetric
	beginTime time.Time
}

func (m *TimerMetric) NewObservation() *TimerMetricObservation {
	τ := new(TimerMetricObservation)
	τ.m = m
	τ.beginTime = time.Now()
	return τ
}

func (τ *TimerMetricObservation) End() {
	Δ := time.Since(τ.beginTime)
	τ.m.c <- metric{
		Type: METRIC_TIMER,
		Name: τ.m.name,
		Value: Δ,
	}
}
