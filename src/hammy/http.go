package hammy

import (
	"fmt"
	"log"
	"time"
	"runtime"
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
	mReqTimer *TimerMetric
	mReceivedValues *CounterMetric
	mCounter200, mCounter400, mCounter500 *CounterMetric
}

// Initialize metric objects
func (h *HttpServer) InitMetrics(metricsNamespace string) {
	h.ms = NewMetricSet(metricsNamespace, 30*time.Second)
	h.mReqTimer = h.ms.NewTimer("requests")
	h.mReceivedValues = h.ms.NewCounter("received_values")
	h.mCounter200 = h.ms.NewCounter("2xx")
	h.mCounter400 = h.ms.NewCounter("4xx")
	h.mCounter500 = h.ms.NewCounter("5xx")
}

func (h *HttpServer) reqStatistics(req *IncomingMessage) {
	var values_received uint64

	for _, hD := range req.Data {
		for _, v := range hD {
			values_received += uint64(len(v))
		}
	}

	h.mReceivedValues.Add(values_received)
}

// Request handler
func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// statistics
	τ := h.mReqTimer.NewObservation()
	defer func() { τ.End() } ()

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		h.mCounter400.Add(1)
		return
	}

	defer func() { runtime.GC() } ()

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
		case "application/x-msgpack":
			dataDecoder = msgpack.NewDecoder(r.Body, nil)
			dataEncoder = msgpack.NewEncoder(w)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
			fmt.Fprintf(w, "Unsupported Content-Type\n")
			h.mCounter400.Add(1)
			return
	}

	var req IncomingMessage
	err := dataDecoder.Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		fmt.Fprintf(w, "%v\n", err);
		h.mCounter400.Add(1)
		return
	}

	h.reqStatistics(&req) // Statistics

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
		h.mCounter500.Add(1)
		return
	}

	h.mCounter200.Add(1)
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
