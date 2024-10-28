package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	// sqllite support
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	// internal packages
	"github.com/slayerjk/http-param-to-db/internal/logging"
	"github.com/slayerjk/http-param-to-db/internal/mailing"
)

// log default path & logs to keep after rotation
const (
	appName           = "http-param-to-db"
	defaultLogPath    = "logs"
	defaultLogsToKeep = 3
	paramName         = "value"
	dbFile            = "data.db"
)

var startTime = time.Now()

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
				log.Printf("empty %s param posted\n", paramName)
				w.Write([]byte("Empty param"))
				return
			}

			// TODO: add check for name regexp, must be(?) "RP\d+"
			paramPosted := fmt.Sprintf("Param posted: %s", paramVal)
			mailing.SendPlainEmailWoAuth("mailing.json", "report", appName, []byte(paramPosted), startTime)
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
				log.Printf("failed to insert %s param into db:\n\t%v\n", paramName, errI)
			}

			paramProcessed := fmt.Sprintf("%s param successfully processed, waiting for next request", paramVal)
			mailing.SendPlainEmailWoAuth("mailing.json", "report", appName, []byte(paramProcessed), startTime)
			log.Println(paramProcessed)
			db.Close()
			return
		}

		log.Printf("No 'name' param in POST")
		w.Write([]byte("No 'name' parameter!\n"))
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
func StartWebServer(address string, mux *http.ServeMux) error {
	registerHanlers()

	if err := http.ListenAndServe(address, mux); err != nil {
		return err
	}

	log.Println("STARTED!")

	return nil
}

func main() {
	// flags
	logDir := flag.String("log-dir", defaultLogPath, "set custom log dir")
	// logsToKeep := flag.Int("keep-logs", defaultLogsToKeep, "set number of logs to keep after rotation")
	httpPort := flag.String("port", "3000", "http server port")
	flag.Parse()

	// logging
	logFile, err := logging.StartLogging(appName, *logDir, 3)
	if err != nil {
		log.Fatalf("failed to start logging:\n\t%s", err)
	}

	defer logFile.Close()

	// starting programm notification
	// startTime := time.Now()
	log.Println("Program Started")

	// main code here

	// starting web server
	mux := http.DefaultServeMux

	log.Printf("Http server is going to be started on port %s", *httpPort)

	if err := StartWebServer(":"+*httpPort, mux); err != nil {
		log.Fatalf("failed to start web server:\n\t%v", err)
	}
}
