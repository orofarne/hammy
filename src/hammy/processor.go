package hammy

import (
	"sync"
	"math/rand"
	"time"
)

type Processor struct {
	App  HammyApp
	Exec Executor

	wg sync.WaitGroup
}

func (p *Processor) Run() {
	for {
		var data []*Data = p.App.DB.Pull()
		if data == nil || len(data) == 0 {
			return
		}
		for _, elem := range data {
			p.wg.Add(1)
			go p.processElem(elem)
		}
		p.wg.Wait()
	}
}

func (p *Processor) processElem(elem *Data) {
	defer p.wg.Done()

	var triggers []string = p.App.CL.Get(elem.M)
	var state *interface{}
	var new_data []*Data

	if len(triggers) > 0 {
		state = p.App.SK.Get(elem.M)
	}
	for _, tr := range triggers {
		var new_data_elem *Data;
		new_data_elem, state = p.Exec.Process(tr, elem, state)
		new_data = append(new_data, new_data_elem)
		// Process next steps
	}

	// Detect collisions and retry
	if p.App.SK.Set(elem.M, state) {
		var new_elem *Data
		for _, new_elem = range new_data {
			p.wg.Add(1)
			go p.processElem(new_elem)
			p.App.DB.Push(new_elem)
		}
	} else {
		 // random sleep
		t := rand.Intn(1000)
		time.Sleep(time.Duration(t) * time.Millisecond)

		p.processElem(elem)
	}
}
