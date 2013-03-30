package hammy

import (
	"fmt"
	"log"
	"time"
	"net/http"
	"encoding/json"
	"github.com/ugorji/go-msgpack"
)

// Http server object
// InitMetric must be called before use
type HttpServer struct{
	// Request handler  object
	RHandler RequestHandler
	// Metrics
	ms *MetricSet
	rTimer *TimerMetric
	counter200, counter400, counter500 *CounterMetric
}

// Initialize metric objects
func (h *HttpServer) InitMetrics(metricsNamespace string) {
	h.ms = NewMetricSet(metricsNamespace, 30*time.Second)
	h.rTimer = h.ms.NewTimer("requests")
	h.counter200 = h.ms.NewCounter("2xx")
	h.counter400 = h.ms.NewCounter("4xx")
	h.counter500 = h.ms.NewCounter("5xx")
}

// Request handler
func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// statistics
	before := time.Now()
	defer func() {
		h.rTimer.Add(time.Since(before))
	}()

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		h.counter400.Add(1)
		return
	}

	contentTypeHeader, headerFound := r.Header["Content-Type"]
	var contentType string
	if headerFound && len(contentTypeHeader) > 0 {
		contentType = contentTypeHeader[0]
	} else {
		contentType = "application/json"
	}

	type DataDecoder interface{
		Decode(interface{}) error
	}

	type DataEncoder interface{
		Encode(interface{}) error
	}

	var dataDecoder DataDecoder
	var dataEncoder DataEncoder
	switch contentType {
		case "application/json":
			dataDecoder = json.NewDecoder(r.Body)
			dataEncoder = json.NewEncoder(w)
		case "application/octet-stream":
			dataDecoder = msgpack.NewDecoder(r.Body, nil)
			dataEncoder = msgpack.NewEncoder(w)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
			fmt.Fprintf(w, "Unsupported Content-Type\n")
			h.counter400.Add(1)
			return
	}

	var req IncomingMessage
	err := dataDecoder.Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		fmt.Fprintf(w, "%v\n", err);
		h.counter400.Add(1)
		return
	}

	errs := h.RHandler.Handle(req.Data)
	errs_str := make(map[string]string)
	for k, e := range errs {
		errs_str[k] = e.Error()
	}

	resp := ResponseMessage{
		Errors: errs_str,
	}

	w.Header().Set("Content-Type", contentType)
	err = dataEncoder.Encode(&resp)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", err);
		log.Printf("Internal Server Error: %v", err)
		h.counter500.Add(1)
		return
	}

	h.counter200.Add(1)
}

// Start http interface and lock goroutine untill fatal error
func StartHttp(rh RequestHandler, cfg Config, metricsNamespace string) error {
	h := &HttpServer{
		RHandler: rh,
	}

	h.InitMetrics(metricsNamespace)

	// Setup server
	s := &http.Server{
		Addr:				cfg.IncomingHttp.Addr,
		Handler:			h,
		ReadTimeout:		30 * time.Second,
		WriteTimeout:		30 * time.Second,
		MaxHeaderBytes:		1 << 20,
	}

	return s.ListenAndServe()
}
