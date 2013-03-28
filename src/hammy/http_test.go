package hammy

import (
	"testing"
	"bytes"
	"net/http"
	"time"
	//"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/ugorji/go-msgpack"
)

type RequestHandlerTestImpl struct {
	Data IncomingData
}

func (rh *RequestHandlerTestImpl) Handle(data IncomingData) (errs map[string]error) {
	errs = make(map[string]error)
	rh.Data = data
	//fmt.Printf("--> Got new data:\n%#v\n\n", data)
	return
}

func TestHttpInterface(t *testing.T) {
	var cfg Config
	cfg.IncomingHttp.Addr = "127.0.0.1:16739"
	SetConfigDefaults(&cfg)
	rh := new(RequestHandlerTestImpl)

	go StartHttp(rh, cfg)
	time.Sleep(100 * time.Millisecond)

	httpAddr := "http://" + cfg.IncomingHttp.Addr + "/"

	// Simple test (Empty but valid data)
	buf := bytes.NewBufferString("{}")
	resp, err := http.Post(httpAddr, "application/json", buf)
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 200 {
		t.Errorf("Expected response of code 200, got: %#v", resp)
	}

	// Simple errors
	resp, err = http.Get(httpAddr)
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 405 {
		t.Errorf("Expected 405, got: %#v", resp)
	}

	buf = bytes.NewBufferString("asdfasdfadsfa")
	resp, err = http.Post(httpAddr, "application/json", buf)
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 400 {
		t.Errorf("Expected response of code 400, got: %#v", resp)
	}

	// Data and checker
	jsonTestData := `{
		"host1": {
			"key1.1": [{
				"timestamp": 1361785778,
				"value": 3.14
			}]
		},
		"host2": {
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

	checkTestData := func() {
		if host1, found := rh.Data["host1"]; !found {
			t.Errorf("`host1` not found")
		} else {
			if key11, found := host1["key1.1"]; !found {
				t.Errorf("`key1.1` not found")
			} else {
				if len(key11) != 1 {
					t.Errorf("Expected len(key11) = 1, got: %d", len(key11))
				} else {
					if key11[0].Timestamp != 1361785778 {
						t.Errorf("Expected %v, got %#v", 1361785778, key11[0].Timestamp)
					}
					val, converted := key11[0].Value.(float64)
					if !converted { t.Errorf("Wrong type %T", key11[0].Value) } else {
						if val != 3.14 {
							t.Errorf("Expected %v, got %#v", 3.14, val)
						}
					}
				}
			}
		}
		if host2, found := rh.Data["host2"]; !found {
			t.Errorf("`host2` not found")
		} else {
			if key21, found := host2["key2.1"]; !found {
				t.Errorf("`key2.1` not found")
			} else {
				if len(key21) != 1 {
					t.Errorf("Expected len(key21) = 1, got: %d", len(key21))
				} else {
					if key21[0].Timestamp != 1361785817 {
						t.Errorf("Expected %v, got %#v", 1361785817, key21[0].Timestamp)
					}
					val, converted := key21[0].Value.(string)
					if !converted { t.Errorf("Wrong type %T", key21[0].Value) } else {
						if val != "test string" {
							t.Errorf("Expected %v, got %#v", "test string", val)
						}
					}
				}
			}
			if key22, found := host2["key2.2"]; !found {
				t.Errorf("`key2.2` not found")
			} else {
				if len(key22) != 2 {
					t.Errorf("Expected len(key22) = 2, got: %d", len(key22))
				} else {
					if key22[0].Timestamp != 1361785858 {
						t.Errorf("Expected %v, got %#v", 1361785858, key22[0].Timestamp)
					}
					val, converted := key22[0].Value.(float64)
					if !converted { t.Errorf("Wrong type %T", key22[0].Value) } else {
						if val != 12345 {
							t.Errorf("Expected %v, got %#v", 12345, val)
						}
					}
					if key22[1].Timestamp != 1361785927 {
						t.Errorf("Expected %v, got %#v", 1361785927, key22[1].Timestamp)
					}
					val, converted = key22[1].Value.(float64)
					if !converted { t.Errorf("Wrong type %T", key22[1].Value) } else {
						if val != 999.3 {
							t.Errorf("Expected %v, got %#v", 999.3, val)
						}
					}
				}
			}
		}
	}

	// JSON
	// Valid
	buf = bytes.NewBufferString(jsonTestData)
	resp, err = http.Post(httpAddr, "application/json", buf)
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { panic("Failed to read body") }
		t.Errorf("Expected response of code 200, got: %#v, body: %s", resp, body)
	}
	checkTestData()

	// Invalid
	buf = bytes.NewBufferString(`{
		"host1": {
			"key1": "booo!"
		}
	}`)
	resp, err = http.Post(httpAddr, "application/json", buf)
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { panic("failed to read body") }
		t.Errorf("Expected response of code 400, got: %#v, body: %s", resp, body)
	}

	// Message Pack
	// Valid
	var buf_data IncomingData
	err = json.Unmarshal([]byte(jsonTestData), &buf_data)
	if err != nil { panic("error unmarshaling test data") }
	enc := msgpack.NewEncoder(buf)
	err = enc.Encode(buf_data)
	if err != nil { panic("error remarshaling (to msgpack) test data") }

	resp, err = http.Post(httpAddr, "application/octet-stream", buf)
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { panic("Failed to read body") }
		t.Errorf("Expected response of code 200, got: %#v, body: %s", resp, body)
	}
	checkTestData()

	// Invalid
	buf = bytes.NewBufferString("adsfasdfasdfasdfasdfasdfasdfasdfasdfasdfasd")
	resp, err = http.Post(httpAddr, "application/octet-stream", buf)
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { panic("failed to read body") }
		t.Errorf("Expected response of code 400, got: %#v, body: %s", resp, body)
	}
}
