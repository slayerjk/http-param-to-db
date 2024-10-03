package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	// change this path for your project
	"api-to-db/internal/logging"
	"api-to-db/internal/rotatefiles"
)

// log default path & logs to keep after rotation
const (
	defaultLogPath    = "logs"
	defaultLogsToKeep = 3
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

func main() {
	// flags
	logDir := flag.String("log-dir", defaultLogPath, "set custom log dir")
	logsToKeep := flag.Int("keep-logs", defaultLogsToKeep, "set number of logs to keep after rotation")
	flag.Parse()

	// logging
	appName := "api-to-db"

	logFile, err := logging.StartLogging(appName, *logDir, 3)
	if err != nil {
		log.Fatalf("failed to start logging:\n\t%s", err)
	}

	defer logFile.Close()

	// starting programm notification
	startTime := time.Now()
	log.Println("Program Started")

	// main code here

	// starting web server
	mux := http.DefaultServeMux
	registerHanlers()
	errMux := http.ListenAndServe(":3000", mux)
	if err != nil {
		log.Fatalf("failed to run server:\n\t%v", errMux)
	}

	// count & print estimated time
	endTime := time.Now()
	log.Printf("Program Done\n\tEstimated time is %f seconds", endTime.Sub(startTime).Seconds())

	// close logfile and rotate logs
	logFile.Close()

	if err := rotatefiles.RotateFilesByMtime(*logDir, *logsToKeep); err != nil {
		log.Fatalf("failed to rotate logs:\n\t%s", err)
	}
}
