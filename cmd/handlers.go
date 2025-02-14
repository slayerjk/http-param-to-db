package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	mailing "github.com/slayerjk/go-mailing"
)

// root http handler
func (app *application) rootHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("Got query", "host", r.Host, "url path", r.URL.Path, "remote addr", r.RemoteAddr, "method", r.Method)
	w.Write([]byte("HELLO!"))
}

// extract query parameter handler
func (app *application) postHandler(w http.ResponseWriter, r *http.Request) {
	// define result var
	var paramVal string

	app.logger.Info("Got query", "host", r.Host, "url path", r.URL.Path, "remote addr", r.RemoteAddr, "method", r.Method)

	// process only POST requests
	if r.Method == "POST" {

		switch app.mode {

		case "param":
			// porcess only request with paramName flag(value of app.paramName) in it
			if !r.URL.Query().Has(app.paramName) {
				errParamNo := fmt.Sprintf("no required param(%s) in body", app.paramName)
				w.Write([]byte(errParamNo))
				app.logger.Warn(errParamNo)
				return
			}

			paramVal = r.URL.Query().Get(app.paramName)

			// skip empty param
			if len(paramVal) == 0 {
				app.logger.Warn("empty param posted", "param", app.paramName)
				w.Write([]byte("empty param"))
				return
			}

			// TODO: add check for name regexp, must be(?) "RP\d+" (data$11101)
			paramPosted := fmt.Sprintf("Param posted: %s", paramVal)
			// mail this error if mailing option is on
			// if app.mailingOpt {
			// mailErr = mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramPosted))
			// if mailErr != nil {
			// 	logger.Printf("failed to send email:\n\t%v", mailErr)
			// }
			// }
			app.logger.Info(paramPosted)
			w.Write([]byte("OK"))

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
				w.Write([]byte("bad request's body"))
				return
			}

			// unmarshall json
			errU := json.Unmarshal(bytesBody, &reqBody)
			if errU != nil {
				app.logger.Warn("failed to unmarshall request body", slog.Any("err", errU))
				w.Write([]byte("bad request's body"))
				return
			}

			app.logger.Info("body posted", "body", string(bytesBody))

			// check if there is map key(and value) of app.jsonParam
			if _, ok := reqBody[app.paramName]; !ok {
				errParamNo := fmt.Sprintf("no required param(%s) in body", app.paramName)
				app.logger.Warn(errParamNo)
				w.Write([]byte(errParamNo))
				return
			}

			// check if param empty
			paramVal = reqBody[app.paramName].(string)
			if len(paramVal) == 0 {
				errParamEmpty := fmt.Sprintf("empty param(%s) in body", app.paramName)
				app.logger.Warn(errParamEmpty)
				w.Write([]byte(errParamEmpty))
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
					w.Write([]byte("OK"))
					return
				}
			}
			w.Write([]byte("OK"))
		}

		// 3 atempts to insert data into db
		for i := 1; i < 4; i++ {
			// open db
			db, err := sql.Open("sqlite3", "file:"+dbFile)
			if err != nil {
				paramDbOpen := fmt.Sprintf("failed to open db before insert:\n\t%v", err)
				app.logger.Error(paramDbOpen)
				// mail this error if mailing option is on
				if app.mailingOpt {
					mailErr = mailing.SendPlainEmailWoAuth(app.mailingFile, "error", appName, []byte(paramDbOpen))
					if mailErr != nil {
						app.logger.Warn("failed to send email", slog.Any("err", mailErr))
					}
				}
				os.Exit(1)
			}
			defer db.Close()
			// insert name param into db
			postedDate := time.Now().Format("02.01.2006 15:04:05")
			query := fmt.Sprintf(
				"INSERT INTO %s (%s, %s) values('%s', '%s')",
				app.dbDataTable,
				app.dbValueColumn,
				app.dbPostedDateColumn,
				paramVal,
				postedDate,
			)
			_, errI := db.Exec(query)

			if errI != nil {
				paramDbInsert := fmt.Sprintf("failed to insert '%s' param into db('%s'):\n\t%v\n", paramVal, dbFile, errI)

				// repeat only if sqlite3 error "sqlite3: unable to open database file"
				regexpErrUnique := regexp.MustCompile("sqlite3: unable to open database file")
				errorStr := fmt.Sprintf("%v", errI)
				if len(regexpErrUnique.Find([]byte(errorStr))) != 0 {
					app.logger.Info(paramDbInsert)
					app.logger.Warn("attemp to insert data into db failed, trying again in 5 sec", slog.Any("attempt", i))
					time.Sleep(5 * time.Second)
					db.Close()
					// stop attempts to insert if it's 3d attempt already
					if i == 3 {
						app.logger.Warn("all 3 attempts is failed")
						app.logger.Info(paramDbInsert)
						// mail this error if mailing option is on
						if app.mailingOpt {
							mailErr = mailing.SendPlainEmailWoAuth(app.mailingFile, "error", appName, []byte(paramDbInsert))
							if mailErr != nil {
								app.logger.Warn("failed to send email", slog.Any("err", mailErr))
							}
						}
						return
					}
					continue
				}

				// if sqlite3 error not "sqlite3: unable to open database file" - return
				db.Close()
				app.logger.Info(paramDbInsert)
				// mail this Derror if mailing option is on
				if app.mailingOpt {
					mailErr = mailing.SendPlainEmailWoAuth(app.mailingFile, "error", appName, []byte(paramDbInsert))
					if mailErr != nil {
						app.logger.Warn("failed to send email", slog.Any("err", mailErr))
					}
				}
				return
			}

			paramProcessed := fmt.Sprintf("%s param successfully processed, waiting for next request", paramVal)
			// mail this if mailing option is on
			if app.mailingOpt {
				mailErr = mailing.SendPlainEmailWoAuth(app.mailingFile, "report", appName, []byte(paramProcessed))
				if mailErr != nil {
					app.logger.Warn("failed to send email", slog.Any("err", mailErr))
				}
			}

			app.logger.Info(paramProcessed)
			db.Close()
			break
		}

		return
	}

	w.Write([]byte("Only POST allowed!\n"))
}
