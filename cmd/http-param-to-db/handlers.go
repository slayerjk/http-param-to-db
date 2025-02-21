package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	mailing "github.com/slayerjk/go-mailing"
)

// GET - root http handler(returns nothing in response) d
func (app *application) rootHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("Got query", "host", r.Host, "url path", r.URL.Path, "remote addr", r.RemoteAddr, "method", r.Method)
	w.WriteHeader(http.StatusOK)
}

// GET - query for /api(NOT FOUND)
func (app *application) apiGetHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// POST - extract query parameter handler
func (app *application) apiPostHandler(w http.ResponseWriter, r *http.Request) {
	var paramVal string

	app.logger.Info("Got query", "host", r.Host, "url path", r.URL.Path, "remote addr", r.RemoteAddr, "method", r.Method)

	switch app.mode {

	case "param":
		// porcess only request with paramName flag(value of app.paramName) in it
		if !r.URL.Query().Has(app.paramName) {
			errParamNo := fmt.Sprintf("no required param(%s) in body", app.paramName)
			http.Error(w, errParamNo, http.StatusBadRequest)
			app.logger.Warn(errParamNo)
			return
		}

		paramVal = r.URL.Query().Get(app.paramName)

		// skip empty param
		if len(paramVal) == 0 {
			app.logger.Warn("empty param posted", "param", app.paramName)
			http.Error(w, "empty param", http.StatusBadRequest)
			return
		}

		paramPosted := fmt.Sprintf("Param posted: %s", paramVal)

		app.logger.Info(paramPosted)

	case "body":
		// define request body
		var (
			reqBody          map[string]any
			bodyConditionKey string
			bodyConditionVal string
		)

		// read request body
		bytesBody, errR := io.ReadAll(r.Body)
		if errR != nil {
			app.logger.Warn("failed to read request body", slog.Any("err", errR))
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		// unmarshall json
		errU := json.Unmarshal(bytesBody, &reqBody)
		if errU != nil {
			app.logger.Warn("failed to unmarshall request body", slog.Any("err", errU))
			http.Error(w, "bad json data", http.StatusBadRequest)
			return
		}

		app.logger.Info("body posted", "body", string(bytesBody))

		// check if there is map key(and value) of app.jsonParam
		if _, ok := reqBody[app.paramName]; !ok {
			errParamNo := fmt.Sprintf("no required param(%s) in body", app.paramName)
			app.logger.Warn(errParamNo)
			http.Error(w, errParamNo, http.StatusBadRequest)
			return
		}

		// check if param empty
		paramVal = reqBody[app.paramName].(string)
		if len(paramVal) == 0 {
			errParamEmpty := fmt.Sprintf("empty param(%s) in body", app.paramName)
			app.logger.Warn(errParamEmpty)
			http.Error(w, "empty param", http.StatusBadRequest)
			return
		}

		// check body condition
		if len(app.bodyCondition) != 0 {
			bodyConditionKey = strings.Split(app.bodyCondition, ":")[0]
			bodyConditionVal = strings.Split(app.bodyCondition, ":")[1]
		}
		// check only if flag is not empty
		if app.bodyCondition != "" {
			if reqBody[bodyConditionKey] != bodyConditionVal {
				app.logger.Warn("additional condition for request body is not met", "condition", app.bodyCondition)
				http.Error(w, "additional condittion is not met", http.StatusBadRequest)
				return
			}
		}
	}

	// insert processed data
	errIns := app.db.InsertProcessed(app.dbFile, app.dbDataTable, app.dbValueColumn, app.dbPostedDateColumn, paramVal)
	if errIns != nil {
		errInsert := fmt.Sprintf("failed to Insert %s into db:\n\t%v", paramVal, errIns)
		app.logger.Error(errInsert)

		// don't mail if error contains "sqlite3: constraint failed: UNIQUE constraint failed"
		uniqueConstMatched, err := regexp.Match("sqlite3: constraint failed: UNIQUE constraint failed", []byte(errIns.Error()))
		if err != nil {
			app.logger.Warn("failed to match insert error with unique constraint condition")
		}
		if uniqueConstMatched {
			http.Error(w, "already processed", http.StatusBadRequest)
			return
		}

		// mail this error if mailing option is on
		if app.mailingOpt {
			mailErr := mailing.SendPlainEmailWoAuth(app.mailingFile, "error", appName, []byte(errInsert))
			if mailErr != nil {
				app.logger.Warn("failed to send email", slog.Any("err", mailErr))
			}
		}

		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// log success
	paramProcessed := fmt.Sprintf("%s param successfully processed, waiting for next request", paramVal)
	// mail this if mailing option is on
	if app.mailingOpt {
		mailErr := mailing.SendPlainEmailWoAuth(app.mailingFile, "report", appName, []byte(paramProcessed))
		if mailErr != nil {
			app.logger.Warn("failed to send email", slog.Any("err", mailErr))
		}
	}

	app.logger.Info(paramProcessed)

	w.WriteHeader(http.StatusAccepted)
}
