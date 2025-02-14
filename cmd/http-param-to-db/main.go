package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	// sqllite support

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	// internal packages

	mailing "github.com/slayerjk/go-mailing"
	vafswork "github.com/slayerjk/go-vafswork"
)

const (
	appName = "HTTP-PARAM-TO-DB"
)

type application struct {
	// mailing option
	mailingOpt bool
	// mailing data file
	mailingFile string
	// logger options
	logger *slog.Logger
	// db file path
	dbFile string
	// db Data table name
	dbDataTable string
	// db column name that contain value
	dbValueColumn string
	// db column name that contain date of post
	dbPostedDateColumn string
	// application mode
	mode string
	// json value to parse in POST request
	paramName string
	// additional body condition('key:value') to parse in POST request
	bodyCondition string
}

func main() {
	// defining default values
	var (
		dbFile             string = vafswork.GetExePath() + "/data/data.db"
		mailingFileDefault string = vafswork.GetExePath() + "/data/mailing.json"
		logPath            string = vafswork.GetExePath() + "/logs" + "_" + appName
		mailErr            error
	)

	// flags
	logsDir := flag.String("log-dir", logPath, "set custom log dir")
	logsToKeep := flag.Int("keep-logs", 30, "set number of logs to keep after rotation")
	httpPort := flag.String("port", ":3000", "http server port, example for localhost:3000 = ':3000'")
	mode := flag.String("mode", "body", "work mode: wait for url 'param' or 'body' contente(json)")
	paramName := flag.String("param-name", "UUID", "param name/json value to process")
	bodyCondition := flag.String("body-condition", "", "additional json 'body' condition to accept, format is 'key:value'")
	mailingOpt := flag.Bool("m", false, "turn the mailing options on(use 'data/mailing.json')")
	mailingFile := flag.String("mailing-file", mailingFileDefault, "full path to 'mailing.json'")

	flag.Parse()

	// logging
	// create log dir
	if err := os.MkdirAll(*logsDir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stdout, "failed to create log dir %s:\n\t%v", *logsDir, err)
		os.Exit(1)
	}
	// set current date
	dateNow := time.Now().Format("02.01.2006")
	// create log file
	logFilePath := fmt.Sprintf("%s/%s_%s.log", *logsDir, appName, dateNow)
	// open log file in append mode
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to open created log file %s:\n\t%v", logFilePath, err)
		os.Exit(1)
	}
	defer logFile.Close()
	// set logger
	logger := slog.New(slog.NewTextHandler(logFile, nil))

	// init application with deps
	app := &application{
		mailingOpt:         *mailingOpt,
		mailingFile:        *mailingFile,
		logger:             logger,
		dbFile:             dbFile,
		dbDataTable:        "Data",
		dbValueColumn:      "Value",
		dbPostedDateColumn: "Posted_Date",
		mode:               *mode,
		paramName:          *paramName,
		bodyCondition:      *bodyCondition,
	}

	// starting programm notification
	// startTime := time.Now()
	logger.Info("Program Started", slog.Any("MODE", *mode))

	// rotate logs first
	logger.Info("logrotate first")
	if err := vafswork.RotateFilesByMtime(*logsDir, *logsToKeep); err != nil {
		logger.Warn("failure to rotate logs", slog.Any("ERR", err))
	}
	logger.Info("logs rotation done")

	// main code here

	// no point to start program if there is no db file
	if _, errDb := os.Stat(dbFile); errDb != nil {
		// mail this error if mailing option is on
		if *mailingOpt {
			mailErr = mailing.SendPlainEmailWoAuth(*mailingFile, "ERR", appName, []byte("cant find 'data/data.db' file"))
			if mailErr != nil {
				logger.Warn("failed to send email", slog.Any("ERR", mailErr))
			}
		}
		logger.Error("db file doesn't exist", slog.Any("DB_FILE", app.dbFile))
		os.Exit(1)
	}

	// Start Web Server
	logger.Info("Http server is going to be started", "PORT", *httpPort)

	errS := http.ListenAndServe(*httpPort, app.routes())
	if errS != nil {
		logger.Error("failed to start web server", slog.Any("ERR", errS))
		os.Exit(1)
	}
}
