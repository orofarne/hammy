package hammy

import (
	"testing"
	"time"
	"net/http"
	_ "expvar"
	"io/ioutil"
	"encoding/json"
)

func TestMetrics(t *testing.T) {
	const httpAddr = "127.0.0.1:16740"

	go func() {
		t.Fatal(http.ListenAndServe(httpAddr, nil))
	}()

	ms := NewMetricSet("metrics_test", 100*time.Millisecond)
	if ms == nil { t.Fatalf("ms is nil") }

	counter := ms.NewCounter("test_counter")
	counter.Add(1)
	counter.Add(2)
	timer := ms.NewTimer("test_timer")
	timer.Add(3*time.Second)
	timer.Add(1*time.Second)

	time.Sleep(150*time.Millisecond)

	resp, err := http.Get("http://" + httpAddr + "/debug/vars")
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got: %#v", resp)
	}
	dec := json.NewDecoder(resp.Body)
	var res map[string]interface{}
	err = dec.Decode(&res)
	if err != nil { t.Fatalf("JSON error: %#v", err) }
	testData := res["metrics_test"].(map[string]interface{})

	//counter
	counterRes, found := testData["test_counter#rps"]
	if !found { t.Fatalf("test_counter#rps not found") }
	if v, ok := counterRes.(float64); !ok || v < 25 || v > 35 {
		t.Errorf("Invalid test_counter#rps value: %#v, expected ~30", counterRes)
	}

	//timer
	timerResRPS, found := testData["test_timer#rps"]
	if !found { t.Fatalf("test_timer#rps not found") }
	if v, ok := timerResRPS.(float64); !ok || v < 15 || v > 25 {
		t.Errorf("Invalid test_timer#rps value: %#v, expected ~20", timerResRPS)
	}
	timerResAVG, found := testData["test_timer_avgtime#s"]
	if !found { t.Fatalf("test_timer_avgtime#s not found") }
	if v, ok := timerResAVG.(float64); !ok || v < 1.5 || v > 2.5 {
		t.Errorf("Invalid test_timer_avgtime#s value: %#v, expected ~2", timerResAVG)
	}

	time.Sleep(100*time.Millisecond)

	resp, err = http.Get("http://" + httpAddr + "/debug/vars")
	if err != nil { t.Fatalf("Got error: %v", err) }
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got: %#v", resp)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil { t.Fatalf("Read error: %#v", err) }
	err = json.Unmarshal(body, &res)
	if err != nil { t.Fatalf("JSON error: %#v, json: %s", err, body) }
	testData = res["metrics_test"].(map[string]interface{})

	//counter
	counterRes, found = testData["test_counter#rps"]
	if !found { t.Fatalf("test_counter#rps not found") }
	if v, ok := counterRes.(float64); !ok || v != 0 {
		t.Errorf("Invalid test_counter#rps value: %#v, expected 0", counterRes)
	}

	//timer
	timerResRPS, found = testData["test_timer#rps"]
	if !found { t.Fatalf("test_timer#rps not found") }
	if v, ok := timerResRPS.(float64); !ok || v != 0 {
		t.Errorf("Invalid test_timer#rps value: %#v, expected 0", timerResRPS)
	}
	timerResAVG, found = testData["test_timer_avgtime#s"]
	if !found { t.Fatalf("test_timer_avgtime#s not found") }
	if v, ok := timerResAVG.(float64); !ok || v != 0 {
		t.Errorf("Invalid test_timer_avgtime#s value: %#v, expected 0", timerResAVG)
	}

}
