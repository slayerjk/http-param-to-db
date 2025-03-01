package main

import "net/http"

// routes method for app, return servemux with registered handlers in handlers.go
func (app *application) routes() http.Handler {
	// starting web server
	mux := http.NewServeMux()

	// Register HTTP handlers
	mux.HandleFunc("GET /", app.rootHandler)
	mux.HandleFunc("GET /api", app.apiGetHandler)
	mux.HandleFunc("POST /api", app.apiPostHandler)

	return app.recoverPanic(app.logRequest(commonHeaders(mux)))
}
