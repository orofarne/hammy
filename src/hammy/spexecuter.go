package hammy

import (
	"fmt"
	"io"
	"os/exec"
	"bytes"
	"syscall"
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
	S *State
	IData IncomingObjectData
}

//Executer implementation for subprocesses with MessagePack-based RPC
type SPExecuter struct {
	CmdLine string
	MaxIter uint
	Workers chan *process
}

//Create new instance of SPExecutor
//per process
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
		data IncomingObjectData) (cmdb *CmdBuffer, err error) {
//
	cmdb = NewCmdBuffer(0)
	//Fetch worker (may be wait for free worker)
	worker, err := e.getWorker()
	if err != nil {
		return
	}
	defer e.freeWorker(worker)

	//marshal and send args
	pInput := WorkerProcessInput{
		Key: key,
		Trigger: trigger,
		S: state,
		IData: data,
	}

	enc := msgpack.NewEncoder(worker.Stdin)
	err = enc.Encode(pInput)
	if err != nil {
		return
	}

	fmt.Printf("Waiting for result...\n")

	//wait, read and unmarshal result
	dec := msgpack.NewDecoder(worker.Stdout, nil)
	err = dec.Decode(cmdb)
	if err != nil {
		err = fmt.Errorf("SPExexuter error: %#v, child stderr: %#v", err, worker.Stderr.String())
	}

	fmt.Printf("Result: %#v\n", cmdb)

	return
}

//Fetch worker (may be wait for free worker)
func (e *SPExecuter) getWorker() (worker *process, err error) {
	worker = <- e.Workers

	if worker == nil {
		panic("nil worker")
	}

	if worker.Cmd == nil {
		//Creating new subprocess
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

//Return worker to buffer
func (e *SPExecuter) freeWorker(worker *process) {
	//Check process state
	var status syscall.WaitStatus
	var rusage syscall.Rusage
	_, err := syscall.Wait4(worker.Process.Pid, &status, 0, &rusage)

	switch {
		case err == nil && !status.Continued():
			worker.Cmd = nil
		case err != nil:
			fallthrough
		case worker.Count >= e.MaxIter: //Check iteration count
			err = worker.Process.Kill()
			_ = err //FIXME
			worker.Cmd = nil
		default:
	}

	e.Workers <- worker
}
