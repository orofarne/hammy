package hammy

import (
	"testing"
	"time"
)

type SendBufferTestImpl struct {
	RHandler RequestHandler
}

func (sb *SendBufferTestImpl) Push(data *IncomingData) {
	sb.RHandler.Handle(*data)
}

func TestCmdBufferSendCommand(t *testing.T) {
	rh := new(RequestHandlerTestImpl)
	sb := SendBufferTestImpl{
		RHandler: rh,
	}

	cbp := CmdBufferProcessorImpl{
		SBuffer: &sb,
	}

	cmdb := make(CmdBuffer, 1)
	cmdb[0].Cmd = "send"
	cmdb[0].Options = make(map[string]interface{})
	cmdb[0].Options["key"] = "key1"
	cmdb[0].Options["value"] = "hello"
	err := cbp.Process("host1", &cmdb)

	if err != nil {
		t.Errorf("Process error: %#v", err)
	}

	if host1, found := rh.Data["host1"]; !found {
		t.Errorf("`host1` not found (data: %v)", rh.Data)
	} else {
		if key1, found := host1["key1"]; !found {
			t.Errorf("`key1` not found (data: %v)", rh.Data)
		} else {
			if len(key1) != 1 {
				t.Errorf("Expected len(key1) = 1, got: %d", len(key1))
			} else {
				if key1[0].Timestamp == 0 || key1[0].Timestamp > uint64(time.Now().Unix()) {
					t.Errorf("Expected 0 <= timesetamp <= %v, got %#v", time.Now().Unix(), key1[0].Timestamp)
				}
				val, converted := key1[0].Value.(string)
				if !converted { t.Errorf("Wrong type %T", key1[0].Value) } else {
					if val != "hello" {
						t.Errorf("Expected %#v, got %#v", "hello", val)
					}
				}
			}
		}
	}

	cmdb[0].Options["host"] = "host2"
	cmdb[0].Options["value"] = "world"
	err = cbp.Process("host1", &cmdb)

	if err != nil {
		t.Errorf("Process error: %#v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if host2, found := rh.Data["host2"]; !found {
		t.Errorf("`host2` not found (data: %v)", rh.Data)
	} else {
		if key1, found := host2["key1"]; !found {
			t.Errorf("`key1` not found (data: %v)", rh.Data)
		} else {
			if len(key1) != 1 {
				t.Errorf("Expected len(key1) = 1, got: %d", len(key1))
			} else {
				if key1[0].Timestamp == 0 || key1[0].Timestamp > uint64(time.Now().Unix()) {
					t.Errorf("Expected 0 <= timesetamp <= %v, got %#v", time.Now().Unix(), key1[0].Timestamp)
				}
				val, converted := key1[0].Value.(string)
				if !converted { t.Errorf("Wrong type %T", key1[0].Value) } else {
					if val != "world" {
						t.Errorf("Expected %#v, got %#v", "world", val)
					}
				}
			}
		}
	}
}
