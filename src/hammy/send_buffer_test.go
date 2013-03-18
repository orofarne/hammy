package hammy

import (
	"testing"
	"fmt"
	"time"
	"encoding/json"
)

func TestSendBufferSimple(t *testing.T) {
	rh := RequestHandlerTestImpl{}
	var cfg Config
	cfg.SendBuffer.SleepTime = 0.1
	sb := NewSendBufferImpl(&rh, cfg)

	json1 := `{
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
			}]
		}
	}`
	json2 := `{
		"object2": {
			"key2.1": [{
				"timestamp": 1361785819,
				"value": "test string 2"
			}],
			"key2.2": [{
				"timestamp": 1361785858,
				"value": 12345
			},
			{
				"timestamp": 1361785927,
				"value": 999.3
			}]
		},
		"object3": {
			"key3.1": [{
				"timestamp": 1361785788,
				"value": 77.0
			}]
		}
	}`

	var data1, data2 IncomingData
	if err := json.Unmarshal([]byte(json1), &data1); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(json2), &data2); err != nil {
		panic(err)
	}

	go sb.Listen()
	sb.Push(&data1)
	sb.Push(&data2)

	fmt.Printf("[send_buffer_test.go] Sleeping...\n")

	time.Sleep(200 * time.Millisecond)

	// Check data
	if obj1, found := rh.Data["object1"]; !found {
		t.Errorf("`object1` not found")
	} else {
		if key11, found := obj1["key1.1"]; !found {
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
	if obj2, found := rh.Data["object2"]; !found {
		t.Errorf("`object2` not found")
	} else {
		if key21, found := obj2["key2.1"]; !found {
			t.Errorf("`key2.1` not found")
		} else {
			if len(key21) != 2 {
				t.Errorf("Expected len(key21) = 2, got: %d", len(key21))
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
				if key21[1].Timestamp != 1361785819 {
					t.Errorf("Expected %v, got %#v", 1361785819, key21[1].Timestamp)
				}
				val, converted = key21[1].Value.(string)
				if !converted { t.Errorf("Wrong type %T", key21[1].Value) } else {
					if val != "test string 2" {
						t.Errorf("Expected %v, got %#v", "test string 2", val)
					}
				}
			}
		}
		if key22, found := obj2["key2.2"]; !found {
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
	if obj3, found := rh.Data["object3"]; !found {
		t.Errorf("`object3` not found")
	} else {
		if key31, found := obj3["key3.1"]; !found {
			t.Errorf("`key3.1` not found")
		} else {
			if len(key31) != 1 {
				t.Errorf("Expected len(key31) = 1, got: %d", len(key31))
			} else {
				if key31[0].Timestamp != 1361785788 {
					t.Errorf("Expected %v, got %#v", 1361785788, key31[0].Timestamp)
				}
				val, converted := key31[0].Value.(float64)
				if !converted { t.Errorf("Wrong type %T", key31[0].Value) } else {
					if val != 77.0 {
						t.Errorf("Expected %v, got %#v", 77.0, val)
					}
				}
			}
		}
	}


}
