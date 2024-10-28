package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	// sqllite support
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	// internal packages
	"github.com/slayerjk/http-param-to-db/internal/logging"
	"github.com/slayerjk/http-param-to-db/internal/mailing"
)

// root http handler
func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HELLO!"))
}

// extract query parameter handler
func postHandler(w http.ResponseWriter, r *http.Request) {
	// process only POST requests
	if r.Method == "POST" {
		log.Printf("Got query: %v", r.URL.String())

		// porcess only request with paramName(value) in it
		if r.URL.Query().Has(paramName) {
			paramVal := r.URL.Query().Get(paramName)

			// skip empty param
			if len(paramVal) == 0 {
				log.Printf("empty '%s' param posted\n", paramName)
				w.Write([]byte("Empty param"))
				return
			}

			// TODO: add check for name regexp, must be(?) "RP\d+"
			paramPosted := fmt.Sprintf("Param posted: %s", paramVal)
			// mail this error
			mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramPosted), time.Now())
			log.Println(paramPosted)
			w.Write([]byte("OK"))

			// open db
			db, err := sql.Open("sqlite3", "file:"+dbFile)
			if err != nil {
				// TODO: add 'error' email
				log.Fatalf("failed to open db:\n\t%v", err)
			}
			defer db.Close()

			// insert name param into db
			_, errI := db.Exec("INSERT INTO Data(Value) values(?)", paramVal)
			if errI != nil {
				// TODO: add 'error' email
				paramDbInsert := fmt.Sprintf("failed to insert '%s' param into db:\n\t%v\n", paramVal, errI)
				// mail this error
				mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramDbInsert), time.Now())
				log.Println(paramDbInsert)
			}

			paramProcessed := fmt.Sprintf("%s param successfully processed, waiting for next request", paramVal)
			// mail this
			mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramProcessed), time.Now())
			log.Println(paramProcessed)
			db.Close()
			return
		}

		log.Printf("No '%s' param in POST", paramName)
		w.Write([]byte("There is no correct parameter!\n"))
		return
	}

	w.Write([]byte("Only POST allowed!\n"))
	log.Printf("wrong parameter in POST: %v\n", r.URL.String())
}

// Register HTTP handlers
func registerHanlers() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/api", postHandler)
}

// Start Web Server
func StartWebServer(address, dbFile string, mux *http.ServeMux) error {
	registerHanlers()

	if err := http.ListenAndServe(address, mux); err != nil {
		return err
	}

	return nil
}

// get full path of Go executable
func getExePath() string {
	// get executable's working dir
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	exePath := filepath.Dir(exe)

	return exePath
}

// log default path & logs to keep after rotation
const (
	appName   = "http-param-to-db"
	paramName = "value"
)

// defining default values
var (
	LogPath     = getExePath() + "/logs" + "_http-param-to-db"
	LogsToKeep  = 3
	dbFile      = getExePath() + "/data/data.db"
	mailingFile = getExePath() + "/data/mailing.json"
)

func main() {
	// flags
	logsDir := flag.String("log-dir", LogPath, "set custom log dir")
	// logsToKeep := flag.Int("keep-logs", defaultLogsToKeep, "set number of logs to keep after rotation")
	httpPort := flag.String("port", "3000", "http server port")
	flag.Parse()

	// logging
	logFile, err := logging.StartLogging(appName, *logsDir, LogsToKeep)
	if err != nil {
		log.Fatalf("failed to start logging:\n\t%s", err)
	}

	defer logFile.Close()

	// starting programm notification
	// startTime := time.Now()
	log.Println("Program Started")

	// main code here

	// no point to start program if there is no db file
	if _, err := os.Stat(dbFile); err != nil {
		// mail this error
		mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte("cant find db file"), time.Now())
		log.Fatalf("db file(%s) doesn't exist", dbFile)
	}

	// starting web server
	mux := http.DefaultServeMux

	log.Printf("Http server is going to be started on port %s", *httpPort)

	if err := StartWebServer(":"+*httpPort, dbFile, mux); err != nil {
		log.Fatalf("failed to start web server:\n\t%v", err)
	}
}
