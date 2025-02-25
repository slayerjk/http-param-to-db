package main

import "net/http"

// return err to logger.Warn and text to response writer(StatusBadRequest)
func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, loggerError string) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Warn(loggerError, "METHOD", method, "URI", uri)
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

// return err to logger.Warn and text to response writer(StatusBadRequest)
func (app *application) serverError(w http.ResponseWriter, r *http.Request, loggerError string) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(loggerError, "METHOD", method, "URI", uri)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
