package hammy

import (
	"fmt"
	"io"
	"time"
	"os"
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
	Timeout time.Duration
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
	e.Timeout = time.Duration(cfg.Workers.Timeout) * time.Second

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
	defer e.freeWorker(worker)
	if err != nil {
		return
	}

	// Set up timeout
	cEnd := make(chan int)
	go e.workerTimeout(worker, cEnd)

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
		cEnd <- 1
		<- cEnd
		return
	}

	// wait, read and unmarshal result
	dec := msgpack.NewDecoder(worker.Stdout, nil)
	err = dec.Decode(&res)
	cEnd <- 1
	toRes := <- cEnd
	switch {
		case toRes == 2:
			err = fmt.Errorf("SPExexuter timeout for host %v", key)
		case err != nil:
			err = fmt.Errorf("SPExexuter error: %#v, child stderr: %#v", err, worker.Stderr.String())
	}
	return
}

// timeout task
func (e *SPExecuter) workerTimeout(worker *process, cEnd chan int) {
	select {
	case <-cEnd:
		cEnd <- 1
		return
	case <-time.After(e.Timeout):
		err := e.workerKill(worker)
		if err != nil {
			log.Printf("%s", err)
		}
		<- cEnd
		cEnd <- 2
		return
	}
	panic("?!")
}

func (e *SPExecuter) workerKill(worker *process) error {
	defer func() {
		worker.Cmd = nil
	}()

	if worker.Cmd == nil || worker.Cmd.Process == nil {
		return nil
	}

	err := worker.Process.Kill()
	switch err {
		case nil:
			//
		case syscall.ECHILD:
			return nil
		default:
			if e, ok := err.(*os.SyscallError); ok && e.Err == syscall.ECHILD {
				return nil
			}
			return fmt.Errorf("SPExecuter: Process.Kill error: %#v", err)
	}

	// Zombies is not good for us...
	_, err = worker.Process.Wait()
	switch err {
		case nil:
			//
		case syscall.ECHILD:
			return nil
		default:
			if e, ok := err.(*os.SyscallError); ok && e.Err == syscall.ECHILD {
				return nil
			}
			return fmt.Errorf("SPExecuter: Process.Wait error: %#v", err)
	}

	return nil
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
			case err == nil && status.Exited() || err == syscall.ECHILD:
				worker.Cmd = nil
			case err != nil:
				if err2, ok := err.(*os.SyscallError); ok && err2.Err == syscall.ECHILD {
					worker.Cmd = nil
				} else {
					log.Printf("SPExecuter: syscall.Wait4 error: %#v", err)
					err = e.workerKill(worker)
					if err != nil {
						log.Printf("%s", err)
					}
				}
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
			worker.Cmd = nil
			return
		}
		worker.Stdout, err = worker.Cmd.StdoutPipe()
		if err != nil {
			worker.Cmd = nil
			return
		}
		worker.Cmd.Stderr = &worker.Stderr
		err = worker.Start()
		if err != nil {
			worker.Cmd = nil
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
		err := e.workerKill(worker)
		if err != nil {
			log.Printf("%s", err)
		}
	}

	// Return worker to the queue
	e.Workers <- worker
}
