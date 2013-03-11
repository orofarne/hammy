package hammy

import (
	"testing"
	"encoding/json"
	"fmt"
)

type TriggersGetterTestImpl struct {
}

func (tg *TriggersGetterTestImpl) MGet(keys []string) (triggers map[string]string, err error) {
	triggers = make(map[string]string)
	for _, key := range keys {
		triggers[key] = fmt.Sprintf("Trigger for key '%s'", key)
	}
	return
}

type StateKeeperTestImpl struct {
	Data map[string]StateKeeperAnswer
}

func (sk *StateKeeperTestImpl) Generate(keys []string) {
	sk.Data = make(map[string]StateKeeperAnswer)
	for i, key := range keys {
		cas := uint64(i + 100)
		sk.Data[key] = StateKeeperAnswer{
			State: State{},
			Cas: &cas,
			Err: nil,
		}
	}
}

func (sk *StateKeeperTestImpl) Get(key string) StateKeeperAnswer {
	fmt.Printf("StateKeeperTestImpl.Get(%#v)\n", key)

	return sk.Data[key]
}

func (sk *StateKeeperTestImpl) MGet(keys []string) (ret map[string]StateKeeperAnswer) {
	fmt.Printf("StateKeeperTestImpl.MGet(%#v)\n", keys)

	ret = make(map[string]StateKeeperAnswer)
	for _, key := range keys {
		ret[key] = sk.Data[key]
	}
	return
}

func (sk *StateKeeperTestImpl) Set(key string, data State, cas *uint64) (retry bool, err error) {
	fmt.Printf("StateKeeperTestImpl.Set(%#v, %#v, %#v)\n", key, data, cas)

	s := sk.Data[key]

	if *s.Cas == 101 {
		(*s.Cas)++
		sk.Data[key] = s
	}

	if *s.Cas != *cas {
		return true, nil
	}
	s.State = data
	(*s.Cas)++
	sk.Data[key] = s
	return false, nil
}

type ExecuterTestImpl struct {
}

func (e *ExecuterTestImpl) ProcessTrigger(key string, trigger string,
	state *State, data IncomingObjectData) (cmdb *CmdBuffer, err error) {
//
	desired_trigger := fmt.Sprintf("Trigger for key '%s'", key)
	if trigger != desired_trigger {
		return NewCmdBuffer(0), fmt.Errorf("Expected trigger is %#v, got %#v",
			desired_trigger, trigger)
	}
	return NewCmdBuffer(0), nil
}

type CmdBufferProcessorTestImpl struct {
}

func (cbp *CmdBufferProcessorTestImpl) Process(key string, cmdb *CmdBuffer) error {
	return nil
}

func TestRequestImplSimple(t *testing.T) {
	sk := StateKeeperTestImpl{}
	sk.Generate([]string{"object1", "object2"})

	rh := RequestHandlerImpl{
		TGetter: &TriggersGetterTestImpl{},
		SKeeper: &sk,
		Exec: &ExecuterTestImpl{},
		CBProcessor: &CmdBufferProcessorTestImpl{},
	}

	var data IncomingData
	jsonTestData := `{
		"object1": {
			"key1.1": [{
				"timestamp": 1361785778,
				"value": 3.14
			}]
		},
		"object2": {
			"key2.1": [{
				"timestamp": 1361785817,
				"value": "test string"
			}],
			"key2.2": [{
				"timestamp": 1361785858,
				"value": 12345
			},
			{
				"timestamp": 1361785927,
				"value": 999.3
			}]
		}
	}`
	if err := json.Unmarshal([]byte(jsonTestData), &data); err != nil {
		panic(err)
	}

	errs := rh.Handle(data)
	for k, err := range errs {
		if err != nil {
			t.Fatalf("Error for key `%s`: %v", k, err)
		}
	}
}
