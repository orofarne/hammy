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
				cmdb := hammy.CmdBuffer{
					{CmdType: "cmd1", Cmd: "Hello",},
					{CmdType: "cmd2", Cmd: "World",},
				}

				dec := msgpack.NewDecoder(os.Stdin, nil)
				enc := msgpack.NewEncoder(os.Stdout)

				if err := dec.Decode(&input); err != nil {
					log.Fatalf("Decode error: %#v", err)
				}

				output := hammy.WorkerProcessOutput{
					S: input.S,
					CB: &cmdb,
				}

				time.Sleep(100 * time.Millisecond)

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

	cmdb, err := e.ProcessTrigger(key, trigger, &state, data)
	if err != nil {
		t.Fatalf("ProcessTrigger error: %#v", err)
	}

	if len(*cmdb) != 2 {
		t.Fatalf("Invalid size of cmdb: %#v", cmdb)
	}

	if (*cmdb)[0].CmdType != "cmd1" || (*cmdb)[0].Cmd != "Hello" ||
		(*cmdb)[1].CmdType != "cmd2" || (*cmdb)[1].Cmd != "World" {
		t.Errorf("Invalid cmdb: %#v", cmdb)
	}
}
