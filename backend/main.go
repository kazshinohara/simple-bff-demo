package main

import (
	"contrib.go.opencensus.io/exporter/stackdriver"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	port    = os.Getenv("PORT")
	version = os.Getenv("VERSION")
	kind    = os.Getenv("KIND")
)

type rootResponse struct {
	Version string `json:"version"` // v1, v2, v3
	Kind    string `json:"kind"`    // backend, backend-b, backend-c
	Message string `json:"message"`
}

func fetchRootResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		_, span := trace.StartSpanWithRemoteParent(ctx, kind, spanContext)
		defer span.End()
		responseBody, err := json.Marshal(&rootResponse{
			Version: version,
			Kind:    kind,
			Message: "Welcome to " + kind + ". ",
		})
		if err != nil {
			log.Printf("could not json.Marshal: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// for Tracing demo
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-type", "application/json")
		w.Write(responseBody)
	}
}

func main() {
	// Set up Tracing
	exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	if err != nil {
		log.Fatal("CloudTrace: ", err)
	}
	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	// Set up Routing and Server
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", fetchRootResponse).Methods("GET")
	var handler http.Handler = router
	handler = &ochttp.Handler{
		Handler:     handler,
		Propagation: &tracecontext.HTTPFormat{},
	}
	er := http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatal("ListenAndServer: ", er)
	}
}
