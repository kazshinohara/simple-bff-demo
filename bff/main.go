package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var port = os.Getenv("PORT")
var version = os.Getenv("VERSION")
var kind = os.Getenv("KIND")
var backendA = os.Getenv("BE_A")
var backendB = os.Getenv("BE_B")
var backendC = os.Getenv("BE_C")

type commonResponse struct {
	Version string `json:"version"` // v1, v2, v3
	Kind string `json:"kind"` // backend, backend-b, backend-c
	Message string `json:"message"`
}

type bffResponse struct {
	BackendAVersion string `json:"backend_a_version"`
	BackendBVersion string `json:"backend_b_version"`
	BackendCVersion string `json:"backend_c_version"`
}

func fetchBackend(target string, path string) *commonResponse {
	var backendRes commonResponse
	client := &http.Client{}
	client.Timeout = time.Second * 5
	req, err := http.NewRequest("GET", target + path, nil)
	if err != nil {
		log.Printf("could not make a new request: %v", err)
		return &backendRes
	}
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
		Kind: kind,
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

func fetchBffResponse(w http.ResponseWriter, r *http.Request){
	backendARes := fetchBackend(backendA, "")
	backendBRes := fetchBackend(backendB, "")
	backendCRes := fetchBackend(backendC, "")

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
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", fetchRootResponse).Methods("GET")
	router.HandleFunc("/bff", fetchBffResponse).Methods("GET")
	err := http.ListenAndServe(":" + port, router)
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}
