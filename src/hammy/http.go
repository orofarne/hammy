package hammy

import (
	"fmt"
	"log"
	"time"
	"expvar"
	"net/http"
	"encoding/json"
	"github.com/ugorji/go-msgpack"
)

// Golbal http request counter
var httpServerCounter *expvar.Int
// 200-code responses
var httpServer200Couner *expvar.Int
// 400-code responses
var httpServer400Couner *expvar.Int
// 500-code responses
var httpServer500Couner *expvar.Int
// Global timer
var httpServerTime *expvar.Float

func init() {
	httpServerCounter = expvar.NewInt("HttpServerCounter")
	httpServer200Couner = expvar.NewInt("HttpServer200Couner")
	httpServer400Couner = expvar.NewInt("HttpServer400Couner")
	httpServer500Couner = expvar.NewInt("HttpServer500Couner")
	httpServerTime = expvar.NewFloat("HttpServerTime")
}

// Http server object
type HttpServer struct{
	// Request handler  object
	RHandler RequestHandler
}

// Request handler
func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// statistics
	httpServerCounter.Add(1)
	before := time.Now()
	defer func() {
		httpServerTime.Add(time.Since(before).Seconds())
	}()

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		httpServer400Couner.Add(1)
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
			httpServer400Couner.Add(1)
			return
	}

	var req IncomingMessage
	err := dataDecoder.Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		fmt.Fprintf(w, "%v\n", err);
		httpServer400Couner.Add(1)
		return
	}

	errs := h.RHandler.Handle(req.Data)

	resp := ResponseMessage{
		Errors: errs,
	}

	w.Header().Set("Content-Type", contentType)
	err = dataEncoder.Encode(&resp)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", err);
		log.Printf("Internal Server Error: %v", err)
		httpServer500Couner.Add(1)
		return
	}

	httpServer200Couner.Add(1)
}

// Start http interface and lock goroutine untill fatal error
func StartHttp(rh RequestHandler, cfg Config) error {
	h := &HttpServer{
		RHandler: rh,
	}

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
