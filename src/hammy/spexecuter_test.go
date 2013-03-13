package hammy

import (
	"testing"
	"os"
	"os/exec"
	"fmt"
)

func createTestProgramm() (string, error) {
	//Source code
	code := `
		package main

		import (
			"github.com/ugorji/go-msgpack"
			"os"
			"hammy"
			"log"
			"time"
		)

		func main() {
			for {
				var input hammy.WorkerProcessInput
				cmd1opt := make(map[string]string)
				cmd2opt := make(map[string]string)
				cmd1opt["message"] = "Hello"
				cmd2opt["message"] = "World"
				cmdb := hammy.CmdBuffer{
					{Cmd: "cmd1", Options: cmd1opt,},
					{Cmd: "cmd2", Options: cmd2opt,},
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
				break
			}
		}
	`

	//Files
	progSourceFile := os.TempDir() + "/hammy_spexecuter_test_subp.go"
	progFile := os.TempDir() + "/hammy_spexecuter_test_subp"

	//Cretate source file
	f, err := os.Create(progSourceFile)
	if err != nil {
		return "", err
	}

	defer func() {
		//_ = os.Remove(progSourceFile)
	}()

	_, err = f.WriteString(code)
	if err != nil {
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	//Compile programm
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

	prog, err := createTestProgramm()
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
	data := IncomingObjectData{}

	newState, cmdb, err := e.ProcessTrigger(key, trigger, &state, data)
	if err != nil {
		t.Fatalf("ProcessTrigger error: %#v", err)
	}
	_ = newState

	if len(*cmdb) != 2 {
		t.Fatalf("Invalid size of cmdb: %#v", cmdb)
	}

	if (*cmdb)[0].Cmd != "cmd1" || (*cmdb)[0].Options["message"] != "Hello" ||
		(*cmdb)[1].Cmd != "cmd2" || (*cmdb)[1].Options["message"] != "World" {
		t.Errorf("Invalid cmdb: %#v", cmdb)
	}
}
