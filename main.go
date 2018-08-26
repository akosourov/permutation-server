package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// max size for clients body - 16mb
const maxBody = 16 * 1024 * 1024

// MaxBodyHandler impements http.Handler interface and limits max body size
type MaxBodyHandler struct {
	mux          http.Handler
	maxBodyBytes int64
}

func (h MaxBodyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	h.mux.ServeHTTP(w, r)
}

func main() {
	log.Print("Starting server")

	perm := Permutations{
		jobs: map[string]chan []int{},
		mu:   new(sync.RWMutex),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/init", perm.initCtrl)
	mux.HandleFunc("/api/v1/next", perm.nextCtrl)

	server := http.Server{
		Addr:         ":8080",
		Handler:      &MaxBodyHandler{mux: mux, maxBodyBytes: maxBody},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func ErrorJSON(code int, msg string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := map[string]interface{}{"error": msg, "success": false}
	respB, _ := json.Marshal(resp)
	w.Write(respB)
}

func SuccessJSON(data []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
