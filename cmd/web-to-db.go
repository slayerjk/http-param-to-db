package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	// change this path for your project
	hw "web-to-db/internal/handle-web"
	"web-to-db/internal/logging"
	"web-to-db/internal/rotatefiles"
)

// log default path & logs to keep after rotation
const (
	defaultLogPath    = "logs"
	defaultLogsToKeep = 3
)

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
	if err := hw.StartWebServer(":3000", mux); err != nil {
		log.Fatalf("failed to start web server:\n\t%v", err)
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
