package hammy

import (
	"fmt"
	"io"
	"time"
	"os"
	"os/exec"
	"bytes"
	"bufio"
	"syscall"
	"log"
	"strings"
	"github.com/ugorji/go-msgpack"
)


type process struct {
	*exec.Cmd
	Count uint
	PStdin io.Writer
	PStdout io.Reader
	PStderr bytes.Buffer
}

type WorkerProcessInput struct {
	Hostname string
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
	cmdLine string
	maxIter uint
	workers chan *process
	timeout time.Duration

	//Metrics
	ms *MetricSet
	mRequest *TimerMetric
	mExecTimer *TimerMetric
	mWorkerWaitTimer *TimerMetric
	mErrors *CounterMetric
	mCreate *TimerMetric
	mKills *CounterMetric
	mTimeouts *CounterMetric
}

// Create new instance of SPExecutor
// per process
func NewSPExecuter(cfg Config, metricNamespace string) *SPExecuter {
	if cfg.Workers.PoolSize < 1 || cfg.Workers.CmdLine == "" {
		panic("Invalid argument")
	}

	e := new(SPExecuter)
	e.cmdLine = cfg.Workers.CmdLine
	e.maxIter = cfg.Workers.MaxIter
	e.workers = make(chan *process, cfg.Workers.PoolSize)
	e.timeout = time.Duration(cfg.Workers.Timeout) * time.Second

	for i := uint(0); i < cfg.Workers.PoolSize; i++ {
		e.workers <- &process{}
	}

	e.ms = NewMetricSet(metricNamespace, 30 * time.Second)
	e.mRequest = e.ms.NewTimer("request")
	e.mExecTimer = e.ms.NewTimer("exec")
	e.mWorkerWaitTimer = e.ms.NewTimer("worker_wait")
	e.mErrors = e.ms.NewCounter("errors")
	e.mCreate = e.ms.NewTimer("create")
	e.mKills = e.ms.NewCounter("kill")
	e.mTimeouts = e.ms.NewCounter("timeouts")

	return e
}

func (e *SPExecuter) ProcessTrigger(key string, trigger string, state *State,
		data IncomingHostData) (newState *State, cmdb *CmdBuffer, err error) {
//
	ζ := e.mRequest.NewObservation()
	defer func() { ζ.End() } ()

	cmdb = NewCmdBuffer(0)
	newState = NewState()
	res := WorkerProcessOutput{
		CmdBuffer: cmdb,
		State: newState,
	}

	defer func() { if err != nil { e.mErrors.Add(1) } } ()

	// Fetch worker (may be wait for free worker)
	worker, err := e.getWorker()
	defer e.freeWorker(worker)
	if err != nil {
		return
	}

	//Setup statistics
	τ := e.mExecTimer.NewObservation()
	defer func() { τ.End() } ()

	// Set up timeout
	cEnd := make(chan int)
	go e.workerTimeout(worker, cEnd)

	// marshal and send args
	pInput := WorkerProcessInput{
		Hostname: key,
		Trigger: trigger,
		State: state,
		IData: data,
	}

	var errDec error
	buf, errEnc := msgpack.Marshal(pInput)
	if errEnc == nil {
		cEnc := make(chan error)
		go func() {
			_, e := worker.PStdin.Write(buf)
			cEnc <- e
		}()

		// wait, read and unmarshal result
		buffer := bufio.NewReader(worker.PStdout)
		dec := msgpack.NewDecoder(buffer, nil)
		errDec = dec.Decode(&res)
		errEnc = <- cEnc
	}

	cEnd <- 1
	toRes := <- cEnd
	switch {
		case toRes == 2 && errEnc == nil && errDec == nil:
			// FIXME
			log.Printf(">_<")
		case toRes == 2:
			err = fmt.Errorf("SPExexuter timeout for host %v, errors: encoding(%v), decoding(%v), child stderr: %#v",
					key, errEnc, errDec, worker.PStderr.String())
		case errEnc != nil || errDec != nil:
			inf := e.workerInfo(worker)
			e2 := e.workerKill(worker)
			err = fmt.Errorf("SPExexuter error: encoding(%v), decoding(%v), child stderr: %#v, additional info: %s, killed (%v)",
					errEnc, errDec, worker.PStderr.String(), inf, e2)
	}

	if err == nil && worker.PStderr.String() != "" {
		log.Printf("Not empty worker stderr: \"%s\"", worker.PStderr.String())
	}

	return
}

// timeout task
func (e *SPExecuter) workerTimeout(worker *process, cEnd chan int) {
	select {
	case <-cEnd:
		cEnd <- 1
		return
	case <-time.After(e.timeout):
		e.mTimeouts.Add(1)
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

func (e *SPExecuter) workerInfo(worker *process) string {
	if worker.Cmd == nil { return "<worker.Cmd == nil>" }

	var status syscall.WaitStatus
	wpid, err := syscall.Wait4(worker.Process.Pid, &status, syscall.WNOHANG, nil)
	if err != nil {
		return fmt.Sprintf("Wait4 error: %v", err)
	}

	_ = wpid
	return fmt.Sprintf("exit code = %v", status.ExitStatus())
}

func (e *SPExecuter) workerKill(worker *process) error {
	defer func() {
		worker.Cmd = nil
	}()

	if worker.Cmd == nil || worker.Cmd.Process == nil {
		return nil
	}

	e.mKills.Add(1)

	err := worker.Process.Kill()
	switch err {
		case nil:
			//
		case syscall.ECHILD:
			return nil
		default:
			if e, ok := err.(*os.SyscallError); ok && (e.Err == syscall.ECHILD || e.Err == syscall.ESRCH) {
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
	//Statistics
	τ := e.mWorkerWaitTimer.NewObservation()
	defer func() { τ.End() } ()

	worker = <- e.workers

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
		ζ := e.mCreate.NewObservation()
		defer func() { ζ.End() } ()

		// Creating new subprocess
		worker.Count = 0
		worker.Cmd = exec.Command(e.cmdLine)
		worker.PStdin, err = worker.Cmd.StdinPipe()
		if err != nil {
			worker.Cmd = nil
			return
		}
		worker.PStdout, err = worker.Cmd.StdoutPipe()
		if err != nil {
			worker.Cmd = nil
			return
		}
		worker.PStderr.Reset()
		worker.Cmd.Stderr = &worker.PStderr
		err = worker.Start()
		if err != nil {
			if strings.Contains(err.Error(), "cannot allocate memory") {
				panic("cannot allocate memory")
			}
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
	if worker.Count >= e.maxIter {
		err := e.workerKill(worker)
		if err != nil {
			log.Printf("%s", err)
		}
	}

	// Return worker to the queue
	e.workers <- worker
}
