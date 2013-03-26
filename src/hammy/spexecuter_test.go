package hammy

import (
	"testing"
	"os"
	"os/exec"
	"fmt"
)

// Source code
var testWorker1 = `
	package main

	import (
		"github.com/ugorji/go-msgpack"
		"os"
		"hammy"
		"log"
		"time"
		"fmt"
	)

	func main() {
		for i := 0; i < 5; i++ {
			var input hammy.WorkerProcessInput
			cmd1opt := make(map[string]interface{})
			cmd2opt := make(map[string]interface{})
			cmd3opt := make(map[string]interface{})
			cmd1opt["message"] = "Hello"
			cmd2opt["message"] = "World"
			cmd3opt["pid"] = fmt.Sprintf("%d", os.Getpid())
			cmdb := hammy.CmdBuffer{
				{Cmd: "cmd1", Options: cmd1opt,},
				{Cmd: "cmd2", Options: cmd2opt,},
				{Cmd: "cmd3", Options: cmd3opt,},
			}

			dec := msgpack.NewDecoder(os.Stdin, nil)
			enc := msgpack.NewEncoder(os.Stdout)

			if err := dec.Decode(&input); err != nil {
				log.Fatalf("Decode error: %#v", err)
			}

			time.Sleep(100 * time.Millisecond)

			output := hammy.WorkerProcessOutput{
				State: input.State,
				CmdBuffer: &cmdb,
			}

			if err := enc.Encode(&output); err != nil {
				log.Fatalf("Encode error: %#v", err)
			}
		}
	}
`

func createTestProgramm(code string) (string, error) {
	// Files
	progSourceFile := os.TempDir() + "/hammy_spexecuter_test_subp.go"
	progFile := os.TempDir() + "/hammy_spexecuter_test_subp"

	// Cretate source file
	f, err := os.Create(progSourceFile)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = os.Remove(progSourceFile)
	}()

	_, err = f.WriteString(code)
	if err != nil {
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	// Compile programm
	err = os.Chdir(os.TempDir())
	if err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-o", progFile, progSourceFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Error (%v): %s", err, out)
		return "", err
	}

	return progFile, nil
}

func TestSPExecuterSimple(t *testing.T) {
	t.Logf("GOPATH = %v", os.Getenv("GOPATH"))

	prog, err := createTestProgramm(testWorker1)
	if err != nil {
		t.Fatalf("Error creating test programm: %#v", err)
	}
	defer func() {
		os.Remove(prog)
	}()

	cfg := Config{}
	cfg.Workers.PoolSize = 1
	cfg.Workers.CmdLine = prog
	cfg.Workers.MaxIter = 100

	e := NewSPExecuter(cfg)

	key := "test1"
	trigger := `sss@^&%GGGkll""''`
	state := State{}
	data := IncomingHostData{}

	newState, cmdb, err := e.ProcessTrigger(key, trigger, &state, data)
	if err != nil {
		t.Fatalf("ProcessTrigger error: %#v", err)
	}
	_ = newState

	if len(*cmdb) != 3 {
		t.Fatalf("Invalid size of cmdb: %#v", cmdb)
	}

	if (*cmdb)[0].Cmd != "cmd1" || (*cmdb)[0].Options["message"] != "Hello" ||
		(*cmdb)[1].Cmd != "cmd2" || (*cmdb)[1].Options["message"] != "World" ||
		(*cmdb)[2].Cmd != "cmd3" || (*cmdb)[2].Options["pid"] == "" {
		t.Errorf("Invalid cmdb: %#v", cmdb)
	}
}


func TestSPExecuterKills(t *testing.T) {
	t.Logf("GOPATH = %v", os.Getenv("GOPATH"))

	prog, err := createTestProgramm(testWorker1)
	if err != nil {
		t.Fatalf("Error creating test programm: %#v", err)
	}
	defer func() {
		os.Remove(prog)
	}()

	cfg := Config{}
	cfg.Workers.PoolSize = 1
	cfg.Workers.CmdLine = prog
	cfg.Workers.MaxIter = 3

	e := NewSPExecuter(cfg)

	prevPid := ""
	pidChanged := false

	for i := 0; i < 5; i++ {
		key := "test1"
		trigger := `sss@^&%GGGkll""''`
		state := State{}
		data := IncomingHostData{}

		newState, cmdb, err := e.ProcessTrigger(key, trigger, &state, data)
		if err != nil {
			t.Fatalf("ProcessTrigger error: %#v", err)
		}
		_ = newState

		if len(*cmdb) != 3 {
			t.Fatalf("Invalid size of cmdb: %#v", cmdb)
		}

		if (*cmdb)[0].Cmd != "cmd1" || (*cmdb)[0].Options["message"] != "Hello" ||
			(*cmdb)[1].Cmd != "cmd2" || (*cmdb)[1].Options["message"] != "World" ||
			(*cmdb)[2].Cmd != "cmd3" || (*cmdb)[2].Options["pid"] == "" {
			t.Errorf("Invalid cmdb: %#v", cmdb)
		}

		newPid := (*cmdb)[2].Options["pid"]
		if prevPid == "" {
			prevPid = newPid.(string)
		} else {
			if newPid != prevPid {
				pidChanged = true
				break
			}
		}
	}

	if !pidChanged {
		t.Errorf("Pid not changed")
	}
}

/*
func TestSPExecuterDeads(t *testing.T) {
	t.Logf("GOPATH = %v", os.Getenv("GOPATH"))

	prog, err := createTestProgramm(testWorker1)
	if err != nil {
		t.Fatalf("Error creating test programm: %#v", err)
	}
	defer func() {
		os.Remove(prog)
	}()

	cfg := Config{}
	cfg.Workers.PoolSize = 1
	cfg.Workers.CmdLine = prog
	cfg.Workers.MaxIter = 100

	e := NewSPExecuter(cfg)

	prevPid := ""
	pidChanged := false

	for i := 0; i < 7; i++ {
		key := "test1"
		trigger := `sss@^&%GGGkll""''`
		state := State{}
		data := IncomingHostData{}

		newState, cmdb, err := e.ProcessTrigger(key, trigger, &state, data)
		if err != nil {
			t.Fatalf("ProcessTrigger error: %#v", err)
		}
		_ = newState

		if len(*cmdb) != 3 {
			t.Fatalf("Invalid size of cmdb: %#v", cmdb)
		}

		if (*cmdb)[0].Cmd != "cmd1" || (*cmdb)[0].Options["message"] != "Hello" ||
			(*cmdb)[1].Cmd != "cmd2" || (*cmdb)[1].Options["message"] != "World" ||
			(*cmdb)[2].Cmd != "cmd3" || (*cmdb)[2].Options["pid"] == "" {
			t.Errorf("Invalid cmdb: %#v", cmdb)
		}

		newPid := (*cmdb)[2].Options["pid"]
		if prevPid == "" {
			prevPid = newPid
		} else {
			if newPid != prevPid {
				pidChanged = true
			}
		}
	}

	if !pidChanged {
		t.Errorf("Pid not changed")
	}
}
*/