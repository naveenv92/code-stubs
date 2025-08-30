package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", sayHello)
	log.Print("Serving on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Print("Received GET request on \"/\"")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello!"))
	default:
		log.Printf("Received forbidden request on \"/\": %v", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
