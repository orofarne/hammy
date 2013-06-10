package hammy

import (
	"net"
	"fmt"
	"time"
	"hash/crc32"
	"github.com/ugorji/go-msgpack"
)

type WorkerInput struct {
	Id       uint64
	Hostname string
	Trigger  string
	State    *State
	IData    IncomingHostData
}

type WorkerOutput struct {
	Id        uint64
	CmdBuffer *CmdBuffer
	State     *State
}

type WorkerAns struct {
	Data *WorkerOutput
	Err  error
}

type WorkerReq struct {
	Data    WorkerInput
	TS      time.Time
	ResChan chan *WorkerAns
}

// Executer implementation for worker daemon with MessagePack-based RPC interface
type WExecuter struct {
	socketPath string
	timeout    time.Duration

	conn       net.Conn
	reqChan    chan *WorkerReq
	rcvChan    chan *WorkerAns
	rTable     map[uint64]*WorkerReq

	//Metrics
	ms         *MetricSet
	mRequest   *TimerMetric
	mErrors    *CounterMetric
	mTimeouts  *CounterMetric
}

// Create new instance of WExecutor
func NewWExecuter(cfg Config, metricNamespace string) *WExecuter {
	if cfg.Worker.SocketPath == "" {
		panic("Invalid argument")
	}

	e := new(WExecuter)

	e.socketPath = cfg.Worker.SocketPath

	e.reqChan = make(chan *WorkerReq)
	e.rcvChan = make(chan *WorkerAns)
	e.rTable = make(map[uint64]*WorkerReq)

	e.ms = NewMetricSet(metricNamespace, 30*time.Second)
	e.mRequest = e.ms.NewTimer("request")
	e.mErrors = e.ms.NewCounter("errors")
	e.mTimeouts = e.ms.NewCounter("timeouts")

	go e.sender()
	go e.reader()

	return e
}

func (e *WExecuter) connect() error {
	var err error
	e.conn, err = net.Dial("unix", e.socketPath)
	if err != nil {
		return err
	}

	go e.receiver()

	return nil
}

func (e *WExecuter) getRequestId(key string, cas uint64) uint64 {
	var l uint32
	var h uint32
	var rid uint64

	h = crc32.ChecksumIEEE([]byte(key))
	l = uint32(cas ^ (cas >> 32))

	rid = uint64(l) | (uint64(h) << 32)

	return rid;
}

func (e *WExecuter) ProcessTrigger(key string, trigger string, state *State, cas uint64,
	data IncomingHostData) (newState *State, cmdb *CmdBuffer, err error) {
	//
	ζ := e.mRequest.NewObservation()
	defer func() { ζ.End() }()

	defer func() {
		if err != nil {
			e.mErrors.Add(1)
		}
	}()

	reqId := e.getRequestId(key, cas)

	wInput := WorkerInput{
		Id:       reqId,
		Hostname: key,
		Trigger:  trigger,
		State:    state,
		IData:    data,
	}

	resChan := make(chan *WorkerAns)

	req := WorkerReq{
		Data:    wInput,
		TS:      time.Now(),
		ResChan: resChan,
	}

	e.reqChan <- &req

	res := <- resChan
	close(resChan)

	if res.Err != nil {
		err = res.Err
		return
	}

	if res.Data.Id != reqId {
		panic("WExecuter: Invalid reqId")
	}

	return res.Data.State, res.Data.CmdBuffer, nil
}

func (e *WExecuter) sender() {
	for req := range e.reqChan {
		var err error

		// Check collision
		if e.rTable[req.Data.Id] != nil {
			req.ResChan <- &WorkerAns{
				Data: nil,
				Err:  fmt.Errorf("Collision on id %v", req.Data.Id),
			}
			continue
		}

		// Check connection
		if e.conn == nil {
			err = e.connect()
			if err != nil {
				req.ResChan <- &WorkerAns{
					Data: nil,
					Err:  err,
				}
				time.Sleep(time.Millisecond);
				continue
			}
		}

		buf, err := msgpack.Marshal(req.Data)
		if err != nil {
			req.ResChan <- &WorkerAns{
				Data: nil,
				Err:  err,
			}
			continue
		}

		_, err = e.conn.Write(buf)
		if err != nil {
			_ = e.conn.Close()
			e.conn = nil
			time.Sleep(time.Millisecond)
			req.ResChan <- &WorkerAns{
				Data: nil,
				Err:  err,
			}
			continue
		}

		e.rTable[req.Data.Id] = req
	}
}

func (e *WExecuter) receiver() {
	for {
		cmdb := NewCmdBuffer(0)
		newState := NewState()
		out := WorkerOutput{
			CmdBuffer: cmdb,
			State:     newState,
		}

		// wait, read and unmarshal result
		dec := msgpack.NewDecoder(e.conn, nil)
		err := dec.Decode(&out)
		res := WorkerAns{
			Data: &out,
			Err:  err,
		}
		e.rcvChan <- &res;
	}
}

func (e *WExecuter) reader() {
	// TODO timeout option
	c := time.Tick(1 * time.Second)
	for {
		select {
		case data := <-e.rcvChan:
			if data.Data == nil {
				panic(data.Err)
			}
			req, found := e.rTable[data.Data.Id];
			if found {
				req.ResChan <- data
			}
		case <-c:
			for k, req := range e.rTable {
				if time.Since(req.TS).Seconds() >= 1.0 {
					req.ResChan <- &WorkerAns{
						Data: nil,
						Err:  fmt.Errorf("Timeout"),
					}
					delete(e.rTable, k)
				}
			}
		}
	}
}
