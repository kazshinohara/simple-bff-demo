package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
)

var (
	port     = os.Getenv("PORT")
	version  = os.Getenv("VERSION")
	kind     = os.Getenv("KIND")
	backendA = os.Getenv("BE_A")
	backendB = os.Getenv("BE_B")
	backendC = os.Getenv("BE_C")
)

type commonResponse struct {
	Version string `json:"version"` // v1, v2, v3
	Kind    string `json:"kind"`    // backend, backend-b, backend-c
	Message string `json:"message"`
}

type bffResponse struct {
	BackendAVersion string `json:"backend_a_version"`
	BackendBVersion string `json:"backend_b_version"`
	BackendCVersion string `json:"backend_c_version"`
}

func fetchBackend(target string, path string, ctx context.Context, span *trace.Span) *commonResponse {
	var backendRes commonResponse
	client := &http.Client{}
	client.Timeout = time.Second * 5
	req, err := http.NewRequest("GET", target+path, nil)
	if err != nil {
		log.Printf("could not make a new request: %v", err)
		return &backendRes
	}

	req = req.WithContext(ctx)
	format := &tracecontext.HTTPFormat{}
	format.SpanContextToRequest(span.SpanContext(), req)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("could not feach backend: %v", err)
		return &backendRes
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("could not read response body: %v", err)
	}
	if err := json.Unmarshal(body, &backendRes); err != nil {
		log.Printf("could not json.Unmarshal: %v", err)
	}
	return &backendRes
}

func fetchRootResponse(w http.ResponseWriter, r *http.Request) {
	responseBody, err := json.Marshal(&commonResponse{
		Version: version,
		Kind:    kind,
		Message: "Welcome to " + kind + " API. ",
	})
	if err != nil {
		log.Printf("could not json.Marshal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.Write(responseBody)
}

func fetchBffResponse(w http.ResponseWriter, r *http.Request) {
	// Create span
	ctx, span := trace.StartSpan(context.Background(), kind)
	defer span.End()

	childCtx, cancel := context.WithTimeout(ctx, 3000*time.Millisecond)
	defer cancel()

	backendARes := fetchBackend(backendA, "", childCtx, span)
	backendBRes := fetchBackend(backendB, "", childCtx, span)
	backendCRes := fetchBackend(backendC, "", childCtx, span)

	rootRes := bffResponse{
		BackendAVersion: backendARes.Version,
		BackendBVersion: backendBRes.Version,
		BackendCVersion: backendCRes.Version,
	}

	responseBody, err := json.Marshal(rootRes)
	if err != nil {
		log.Printf("could not json.Marshal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBody)
}

func main() {
	// Set up Tracing
	exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	if err != nil {
		log.Fatal("Tracing: ", err)
	}
	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	// Set up Routing and Server
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", fetchRootResponse).Methods("GET")
	router.HandleFunc("/bff", fetchBffResponse).Methods("GET")
	var handler http.Handler = router
	handler = &ochttp.Handler{
		Handler:     handler,
		Propagation: &tracecontext.HTTPFormat{},
	}
	er := http.ListenAndServe(":"+port, handler)
	if er != nil {
		log.Fatal("ListenAndServer: ", er)
	}
}
