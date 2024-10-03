package handleweb

import (
	"log"
	"net/http"
)

// http handlers
func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HELLO!"))
}

// extract query parameter
func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Printf("Got query: %v", r.URL.String())

		if r.URL.Query().Has("name") {
			log.Printf("Name: %s", r.URL.Query().Get("name"))
			w.Write([]byte("OK"))
			return
		}

		w.Write([]byte("No `name` parameter!\n"))
	}

	w.Write([]byte("Only POST allowed!\n"))
	log.Printf("wrong parameter in POST: %v\n", r.URL.String())
}

func registerHanlers() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/body", postHandler)
}

func StartWebServer(address string, mux *http.ServeMux) error {
	registerHanlers()

	if err := http.ListenAndServe(address, mux); err != nil {
		return err
	}

	return nil
}
