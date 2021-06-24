package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var port = os.Getenv("PORT")
var version = os.Getenv("VERSION")
var kind = os.Getenv("KIND")

type rootResponse struct {
	Version string `json:"version"` // v1, v2, v3
	Kind string `json:"kind"` // backend, backend-b, backend-c
	Message string `json:"message"`
}

func fetchRootResponse(w http.ResponseWriter, r *http.Request)	{
	responseBody, err := json.Marshal(&rootResponse{
		Version: version,
		Kind: kind,
		Message: "Welcome to " + kind + ". ",
	})
	if err != nil {
		log.Printf("could not json.Marshal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.Write(responseBody)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", fetchRootResponse).Methods("GET")
	err := http.ListenAndServe(":" + port, router)
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}