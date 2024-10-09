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
			name := r.URL.Query().Get("name")

			if len(name) == 0 {
				log.Println("empty 'name' param posted")
				w.Write([]byte("Empty name param"))
				return
			}
			// TODO: add check for name regexp, must be(?) "RP\d+"

			log.Printf("Name posted: %s", name)
			w.Write([]byte("OK"))
			return
		}

		log.Printf("No 'name' param in POST")
		w.Write([]byte("No 'name' parameter!\n"))
		return
	}

	w.Write([]byte("Only POST allowed!\n"))
	log.Printf("wrong parameter in POST: %v\n", r.URL.String())
}

func registerHanlers() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/api", postHandler)
}

func StartWebServer(address string, mux *http.ServeMux) error {
	registerHanlers()

	if err := http.ListenAndServe(address, mux); err != nil {
		return err
	}

	log.Println("STARTED!")

	return nil
}
