package hammy

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"github.com/ugorji/go-msgpack"
)

//Request handler object
type HttpServer struct{
	RHandler RequestHandler
}

//Request handler
func (h HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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

	var dataDecoder DataDecoder
	switch contentType {
		case "application/json":
			dataDecoder = json.NewDecoder(r.Body)
		case "application/octet-stream":
			dataDecoder = msgpack.NewDecoder(r.Body, nil)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
			fmt.Fprintf(w, "Unsupported Content-Type\n")
			return
	}

	var data IncomingData
	err := dataDecoder.Decode(&data)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		fmt.Fprintf(w, "%v\n", err);
		return
	}

	errs := h.RHandler.Handle(data)
	if len(errs) > 0 {
		//TODO: correct answer to client
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", errs);
		log.Printf("Internal Server Error: %v", errs)
		return
	}

	fmt.Fprint(w, "ok\n")
}

//Start http interface and lock goroutine untill fatal error
func StartHttp(rh RequestHandler, cfg Config) error {
	h := &HttpServer{
		RHandler: rh,
	}

	//Setup server
	s := &http.Server{
		Addr:				cfg.Http.Addr,
		Handler:			h,
		ReadTimeout:		30 * time.Second,
		WriteTimeout:		30 * time.Second,
		MaxHeaderBytes:		1 << 20,
	}

	return s.ListenAndServe()
}
