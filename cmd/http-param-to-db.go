package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	// sqllite support
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	// internal packages
	"github.com/slayerjk/http-param-to-db/internal/logging"
	"github.com/slayerjk/http-param-to-db/internal/mailing"
	"github.com/slayerjk/http-param-to-db/internal/vafswork"
)

// log default path & logs to keep after rotation
const (
	appName   = "http-param-to-db"
	paramName = "value"
)

// set body struct
type Request struct {
	UUID string `json:"UUID"`
}

// TODO: refactor mails
func main() {
	// defining default values
	var (
		logPath            string = vafswork.GetExePath() + "/logs" + "_http-param-to-db"
		dbFile             string = vafswork.GetExePath() + "/data/data.db"
		mailingFile        string = vafswork.GetExePath() + "/data/mailing.json"
		dbDataTable        string = "Data"
		dbValueColumn      string = "Value"
		dbPostedDateColumn string = "Posted_Date"
		// bodyValue          string = "UUID"
	)

	// flags
	logsDir := flag.String("log-dir", logPath, "set custom log dir")
	logsToKeep := flag.Int("keep-logs", 7, "set number of logs to keep after rotation")
	httpPort := flag.String("port", "3000", "http server port")
	mode := flag.String("mode", "body", "work mode: wait for url 'param' or 'body' contente(json)")
	flag.Parse()

	// logging
	logFile, err := logging.StartLogging(appName, *logsDir, *logsToKeep)
	if err != nil {
		log.Fatalf("failed to start logging:\n\t%s", err)
	}

	defer logFile.Close()

	// starting programm notification
	// startTime := time.Now()
	log.Println("Program Started")
	log.Printf("mode is: %s\n", *mode)

	// main code here

	// no point to start program if there is no db file
	if _, err := os.Stat(dbFile); err != nil {
		// mail this error
		mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte("cant find db file"))
		log.Fatalf("db file(%s) doesn't exist", dbFile)
	}

	// root http handler
	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got query: %s%s from %s, method: %s", r.Host, r.URL.Path, r.RemoteAddr, r.Method)
		w.Write([]byte("HELLO!"))
	}

	// extract query parameter handler
	postHandler := func(w http.ResponseWriter, r *http.Request) {
		// define result var
		var paramVal string
		log.Printf("Got query: %s%s from %s, method: %s", r.Host, r.URL.Path, r.RemoteAddr, r.Method)

		// process only POST requests
		if r.Method == "POST" {

			switch *mode {

			case "param":
				// porcess only request with paramName(value) in it
				if !r.URL.Query().Has(paramName) {
					w.Write([]byte("no param in POST"))
					log.Printf("No '%s' param in POST", paramName)
					return
				}

				paramVal = r.URL.Query().Get(paramName)

				// skip empty param
				if len(paramVal) == 0 {
					log.Printf("empty '%s' param posted\n", paramName)
					w.Write([]byte("empty param"))
					return
				}

				// TODO: add check for name regexp, must be(?) "RP\d+" (data$11101)
				paramPosted := fmt.Sprintf("Param posted: %s", paramVal)
				// mail this error
				// mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramPosted))
				log.Println(paramPosted)
				w.Write([]byte("OK"))

			case "body":
				// define request body struct
				var reqBody Request

				// read request body
				bytesBody, errR := io.ReadAll(r.Body)
				if errR != nil {
					log.Printf("failed to read request body:\n\t%v\n", errR)
					w.Write([]byte("bad request's body"))
					return
				}
				log.Printf("body posted is:\n\t%v", string(bytesBody))

				// unmarshall json
				errU := json.Unmarshal(bytesBody, &reqBody)
				if errU != nil {
					log.Printf("failed to unmarshall request body:\n\t%v\n", errU)
					w.Write([]byte("bad request's body"))
					return
				}

				if len(reqBody.UUID) == 0 {
					log.Println("empty param in body")
					w.Write([]byte("empty param in body"))
					return
				}
				paramVal = reqBody.UUID
				w.Write([]byte("OK"))
			}

			// open db
			db, err := sql.Open("sqlite3", "file:"+dbFile)
			if err != nil {
				// TODO: add 'error' email
				log.Fatalf("failed to open db:\n\t%v", err)
			}
			defer db.Close()

			// insert name param into db
			postedDate := time.Now().Format("02.01.2006 15:04:05")
			query := fmt.Sprintf("INSERT INTO %s (%s, %s) values('%s', '%s')", dbDataTable, dbValueColumn, dbPostedDateColumn, paramVal, postedDate)
			_, errI := db.Exec(query)
			if errI != nil {
				paramDbInsert := fmt.Sprintf("failed to insert '%s' param into db:\n\t%v\n", paramVal, errI)
				// mail this error
				mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramDbInsert))
				log.Println(paramDbInsert)
			}

			paramProcessed := fmt.Sprintf("%s param successfully processed, waiting for next request", paramVal)
			// mail this
			mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramProcessed))
			log.Println(paramProcessed)
			db.Close()
			return
		}

		w.Write([]byte("Only POST allowed!\n"))
	}

	// starting web server
	mux := http.DefaultServeMux

	// Register HTTP handlers
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api", postHandler)

	// Start Web Server
	errS := http.ListenAndServe(":"+*httpPort, mux)
	if errS != nil {
		log.Fatal("failed to start web server")
	}
	log.Printf("Http server is going to be started on port %s", *httpPort)
}
