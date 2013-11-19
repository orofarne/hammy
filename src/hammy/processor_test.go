package hammy

import (
	. "launchpad.net/gocheck"
)

type ProcessorTestSuite struct{}

var _ = Suite(&ProcessorTestSuite{})

type fakeExecutor struct {
}

func (e *fakeExecutor) Process(trigger string, data *Data, state *interface{}) (*Data, *interface{}) {
	return data, state
}

type fakeDataBus struct {
	q    chan []*Data
	stop chan int
}

func newFakeDataBus() *fakeDataBus {
	fdb := fakeDataBus{}
	fdb.q = make(chan []*Data, 100)
	fdb.stop = make(chan int)
	return &fdb
}

func (db *fakeDataBus) Push(d *Data) {
	if d == nil {
		db.q <- []*Data{}
	} else {
		db.q <- []*Data{d}
	}
}

func (db *fakeDataBus) Pull() []*Data {
	select {
	case res := <-db.q:
		return res
	case <-db.stop:
		return []*Data{}
	}
}

func (db *fakeDataBus) Stop() {
	db.stop <- 1
}

type fakeStateKeeper struct {
	m map[string]*interface{}
}

func newFakeStateKeeper() *fakeStateKeeper {
	res := fakeStateKeeper{}
	res.m = make(map[string]*interface{})
	return &res
}

func (sk *fakeStateKeeper) Set(m Metric, data *interface{}) bool {
	var key string
	for k, v := range m {
		key = key + "@" + k + "::" + v
	}
	sk.m[key] = data
	return true
}
func (sk *fakeStateKeeper) Get(m Metric) *interface{} {
	var key string
	for k, v := range m {
		key = key + "@" + k + "::" + v
	}
	res, ok := sk.m[key]
	if !ok {
		return nil
	} else {
		return res
	}
}

type fakeCodeLoader struct {
}

func (cl *fakeCodeLoader) Get(m Metric) []string {
	return []string{"", ""}
}

func (s *ProcessorTestSuite) TestProcessor1(c *C) {
	e := fakeExecutor{}

	app := HammyApp{
		DB: newFakeDataBus(),
		SK: newFakeStateKeeper(),
		CL: &fakeCodeLoader{},
	}

	p := Processor{
		App:  app,
		Exec: &e,
	}

	/*
		m := make(Metric)
		m["host"] = "MyHost"
		m["item"] = "ping"
		app.DB.Push(&Data{
			M: m,
			D: 1,
		})
	*/
	app.DB.Push(nil)

	p.Run()
}
