package hammy

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type Processor struct {
	App  HammyApp
	Exec Executor

	wg sync.WaitGroup
}

func (p *Processor) Run() {
	for {
		var data []Data = p.App.DB.Pull()
		if data == nil || len(data) == 0 {
			return
		}
		indexes := p.groupData(data)
		for h, inds := range indexes {
			p.wg.Add(1)
			go p.processHost(h, data, inds)
		}
		p.wg.Wait()
	}
}

func (p *Processor) groupData(data []Data) map[string][]int {
	res := make(map[string][]int)
	for i, elem := range data {
		indexes, ok := res[elem.Host]
		if !ok {
			res[elem.Host] = []int{i}
		} else {
			res[elem.Host] = append(indexes, i)
		}
	}
	return res
}

func (p *Processor) processHost(host string, data []Data, indexes []int) {
	defer p.wg.Done()

	var new_data []Data

	var htr HostTrigger = p.App.CL.Get(host)
	if len(htr) == 0 {
		return
	}

	var state HostState = p.App.SK.Get(host)

	for _, k := range indexes {
		var d Data = data[k]

		nd, err := p.processItem(d, htr, state)
		if err != nil {
			log.Printf("processItem error: %v", err)
			continue
		}
		for _, e := range nd {
			new_data = append(new_data, e)
		}
	}

	if len(state) != 0 {
		// Detect collisions and retry
		if p.App.SK.Set(host, state) {
			p.App.DB.Push(new_data)
		} else {
			// random sleep
			t := rand.Intn(1000)
			time.Sleep(time.Duration(t) * time.Millisecond)

			p.processHost(host, data, indexes)
		}

	} else {
		if len(new_data) > 0 {
			p.App.DB.Push(new_data)
		}
	}
}

func (p *Processor) processItem(data Data, htr HostTrigger, state HostState) (new_data []Data, err error) {
	new_data = make([]Data, 0)
	var freshness time.Duration
	if !data.Timestamp.IsZero() {
		freshness = time.Since(data.Timestamp)
	} else {
		freshness = 0
	}

	for it, tr := range htr {
		for _, dep := range tr.Dependencies {
			if dep.Item == data.Item && freshness < dep.Timeout {
				d, s := p.Exec.Process(tr.Code, data, state[it])
				if d.Value != nil {
					d.Host = data.Host
					d.Item = it
					new_data = append(new_data, d)
				}
				if s != nil {
					state[it] = s
				} else {
					delete(state, it)
				}
			}
		}
	}

	return
}
