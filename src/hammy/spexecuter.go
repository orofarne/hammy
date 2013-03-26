package hammy

import (
	"fmt"
	"io"
	"os/exec"
	"bytes"
	"syscall"
	"log"
	"github.com/ugorji/go-msgpack"
)


type process struct {
	*exec.Cmd
	Count uint
	Stdin io.Writer
	Stdout io.Reader
	Stderr bytes.Buffer
}

type WorkerProcessInput struct {
	Key string
	Trigger string
	State *State
	IData IncomingHostData
}

type WorkerProcessOutput struct {
	CmdBuffer *CmdBuffer
	State *State
}

// Executer implementation for subprocesses with MessagePack-based RPC
type SPExecuter struct {
	CmdLine string
	MaxIter uint
	Workers chan *process
}

// Create new instance of SPExecutor
// per process
func NewSPExecuter(cfg Config) *SPExecuter {
	if cfg.Workers.PoolSize < 1 || cfg.Workers.CmdLine == "" {
		panic("Invalid argument")
	}

	e := new(SPExecuter)
	e.CmdLine = cfg.Workers.CmdLine
	e.MaxIter = cfg.Workers.MaxIter
	e.Workers = make(chan *process, cfg.Workers.PoolSize)

	for i := uint(0); i < cfg.Workers.PoolSize; i++ {
		e.Workers <- &process{}
	}

	return e
}

func (e *SPExecuter) ProcessTrigger(key string, trigger string, state *State,
		data IncomingHostData) (newState *State, cmdb *CmdBuffer, err error) {
//
	cmdb = NewCmdBuffer(0)
	newState = NewState()
	res := WorkerProcessOutput{
		CmdBuffer: cmdb,
		State: newState,
	}

	// Fetch worker (may be wait for free worker)
	worker, err := e.getWorker()
	if err != nil {
		return
	}
	defer e.freeWorker(worker)

	// marshal and send args
	pInput := WorkerProcessInput{
		Key: key,
		Trigger: trigger,
		State: state,
		IData: data,
	}

	enc := msgpack.NewEncoder(worker.Stdin)
	err = enc.Encode(pInput)
	if err != nil {
		return
	}

	// wait, read and unmarshal result
	dec := msgpack.NewDecoder(worker.Stdout, nil)
	err = dec.Decode(&res)
	if err != nil {
		err = fmt.Errorf("SPExexuter error: %#v, child stderr: %#v", err, worker.Stderr.String())
	}

	return
}

// Fetch worker (may be wait for free worker)
func (e *SPExecuter) getWorker() (worker *process, err error) {
	worker = <- e.Workers

	if worker == nil {
		panic("nil worker")
	}

	if worker.Cmd != nil {
		// Check process state
		var status syscall.WaitStatus

		// We can't use worker.ProcessState (it's available only after a call to Wait or Run)
		wpid, err := syscall.Wait4(worker.Process.Pid, &status, syscall.WNOHANG, nil)

		switch {
			case err == nil && wpid == 0:
				// Do nothing
			case err == nil && status.Exited():
				worker.Cmd = nil
			case err != nil:
				log.Printf("SPExecuter: syscall.Wait4 error: %#v", err)
				err = worker.Process.Kill()
				if err != nil {
					log.Printf("SPExecuter: Process.Kill error: %#v", err)
				}
				worker.Cmd = nil
			default:
				// Do nothing
		}
	}

	if worker.Cmd == nil {
		// Creating new subprocess
		worker.Count = 0
		worker.Cmd = exec.Command(e.CmdLine)
		worker.Stdin, err = worker.Cmd.StdinPipe()
		if err != nil {
			return
		}
		worker.Stdout, err = worker.Cmd.StdoutPipe()
		if err != nil {
			return
		}
		worker.Cmd.Stderr = &worker.Stderr
		err = worker.Start()
		if err != nil {
			return
		}
	}

	return
}

// Return worker to buffer
func (e *SPExecuter) freeWorker(worker *process) {
	// Increment count of execution for the worker
	worker.Count++

	// Check iteration count
	if worker.Count >= e.MaxIter {
		err := worker.Process.Kill()
		if err != nil {
			log.Printf("SPExecuter: Process.Kill error: %#v", err)
		}
		worker.Cmd = nil
	}

	// Return worker to the queue
	e.Workers <- worker
}
