package main

import "net/http"

// routes method for app, return servemux with registered handlers in handlers.go
func (app *application) routes() *http.ServeMux {
	// starting web server
	mux := http.NewServeMux()

	// Register HTTP handlers
	http.HandleFunc("GET /", app.rootHandler)
	http.HandleFunc("POST /api", app.postHandler)

	return mux
}
