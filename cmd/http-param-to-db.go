package main

import (
	"flag"
	"log"
	"net/http"

	// change this path for your project
	hw "http-param-to-db/internal/handle-web"
	"http-param-to-db/internal/logging"
)

// log default path & logs to keep after rotation
const (
	defaultLogPath    = "logs"
	defaultLogsToKeep = 3
)

func main() {
	// flags
	logDir := flag.String("log-dir", defaultLogPath, "set custom log dir")
	// logsToKeep := flag.Int("keep-logs", defaultLogsToKeep, "set number of logs to keep after rotation")
	httpPort := flag.String("port", "3000", "http server port")
	flag.Parse()

	// logging
	appName := "http-param-to-db"

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

	if err := hw.StartWebServer(":"+*httpPort, mux); err != nil {
		log.Fatalf("failed to start web server:\n\t%v", err)
	}

	// // count & print estimated time
	// endTime := time.Now()
	// log.Printf("Program Done\n\tEstimated time is %f seconds", endTime.Sub(startTime).Seconds())

	// // close logfile and rotate logs
	// logFile.Close()

	// if err := rotatefiles.RotateFilesByMtime(*logDir, *logsToKeep); err != nil {
	// 	log.Fatalf("failed to rotate logs:\n\t%s", err)
	// }
}
