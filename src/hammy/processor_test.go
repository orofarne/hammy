package hammy

import (
	. "launchpad.net/gocheck"
	"math"
	"time"
)

type ProcessorTestSuite struct{}

var _ = Suite(&ProcessorTestSuite{})

type fakeExecutor struct {
}

func (e *fakeExecutor) Process(trigger string, data Data, state State) (Data, State) {
	return data, state
}

type fakeDataBus struct {
	q    chan []Data
	stop chan int
}

func newFakeDataBus() *fakeDataBus {
	fdb := fakeDataBus{}
	fdb.q = make(chan []Data, 100)
	fdb.stop = make(chan int)
	return &fdb
}

func (db *fakeDataBus) Push(d []Data) {
	db.q <- d
}

func (db *fakeDataBus) Pull() []Data {
	select {
	case res := <-db.q:
		return res
	case <-db.stop:
		return []Data{}
	}
}

func (db *fakeDataBus) Stop() {
	db.stop <- 1
}

type fakeStateKeeper struct {
	m map[string]HostState
}

func newFakeStateKeeper() *fakeStateKeeper {
	res := fakeStateKeeper{}
	res.m = make(map[string]HostState)
	return &res
}

func (sk *fakeStateKeeper) Set(host string, state HostState) bool {
	sk.m[host] = state
	return true
}
func (sk *fakeStateKeeper) Get(host string) HostState {
	return sk.m[host]
}

type fakeCodeLoader struct {
	Triggers map[string]HostTrigger
}

func newFakeCodeLoader() *fakeCodeLoader {
	cl := fakeCodeLoader{}
	cl.Triggers = make(map[string]HostTrigger)
	return &cl
}

func (cl *fakeCodeLoader) Get(host string) HostTrigger {
	return cl.Triggers[host]
}

func (s *ProcessorTestSuite) TestProcessor1(c *C) {
	e := fakeExecutor{}

	fcl := newFakeCodeLoader()
	htr := NewHostTrigger()
	htr["ping2"] = Trigger{
		Code: "blah-blah-blah...",
		Dependencies: []Dependence{
			Dependence{
				Item:    "ping",
				Timeout: time.Minute,
			},
		},
	}
	fcl.Triggers["MyHost"] = htr

	app := HammyApp{
		DB: newFakeDataBus(),
		SK: newFakeStateKeeper(),
		CL: fcl,
	}

	p := Processor{
		App:  app,
		Exec: &e,
	}

	d := Data{
		Metric: Metric{
			Host: "MyHost",
			Item: "ping",
		},
		Timestamp: time.Now(),
		Value:     math.Pi,
	}
	app.DB.Push([]Data{d})

	app.DB.Push(nil)

	p.Run()

	d2 := app.DB.Pull()
	c.Check(len(d2), Equals, 1)
	c.Check(d2[0].Host, Equals, "MyHost")
	c.Check(d2[0].Item, Equals, "ping2")
	c.Check(d2[0].Value, Equals, math.Pi)
}
