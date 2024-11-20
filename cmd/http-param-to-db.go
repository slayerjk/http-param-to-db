package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
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

const (
	appName = "HTTP-PARAM-TO-DB"
)

func main() {
	// defining default values
	var (
		dbFile             string = vafswork.GetExePath() + "/data/data.db"
		mailingFile        string = vafswork.GetExePath() + "/data/mailing.json"
		logPath            string = vafswork.GetExePath() + "/logs" + "_" + appName
		dbDataTable        string = "Data"
		dbValueColumn      string = "Value"
		dbPostedDateColumn string = "Posted_Date"
		logsToKeep         int    = 7
		mailErr            error
	)

	// flags
	logsDir := flag.String("log-dir", logPath, "set custom log dir")
	// logsToKeep := flag.Int("keep-logs", 7, "set number of logs to keep after rotation")
	httpPort := flag.String("port", "3000", "http server port")
	mode := flag.String("mode", "body", "work mode: wait for url 'param' or 'body' contente(json)")
	paramName := flag.String("param-name", "UUID", "param name/json value to process")
	bodyCondition := flag.String("body-condition", "", "additional json 'body' condition to accept, format is 'key:value'")
	mailingOpt := flag.Bool("m", false, "turn the mailing options on(use 'data/mailing.json')")

	flag.Parse()

	// logging
	logFile, err := logging.StartLogging(appName, *logsDir, logsToKeep)
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
	if _, errDb := os.Stat(dbFile); errDb != nil {
		// mail this error if mailing option is on
		if *mailingOpt {
			mailErr = mailing.SendPlainEmailWoAuth(mailingFile, "error", appName, []byte("cant find 'data/data.db' file"))
			if mailErr != nil {
				log.Printf("failed to send email:\n\t%v", mailErr)
			}
		}

		log.Fatalf("'data/data.db' file(%s) doesn't exist", dbFile)
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
				if !r.URL.Query().Has(*paramName) {
					errParamNo := fmt.Sprintf("no required param(%s) in body", *paramName)
					w.Write([]byte(errParamNo))
					log.Println(errParamNo)
					return
				}

				paramVal = r.URL.Query().Get(*paramName)

				// skip empty param
				if len(paramVal) == 0 {
					log.Printf("empty '%s' param posted\n", *paramName)
					w.Write([]byte("empty param"))
					return
				}

				// TODO: add check for name regexp, must be(?) "RP\d+" (data$11101)
				paramPosted := fmt.Sprintf("Param posted: %s", paramVal)
				// mail this error if mailing option is on
				// if *mailingOpt {
				// mailErr = mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramPosted))
				// if mailErr != nil {
				// 	log.Printf("failed to send email:\n\t%v", mailErr)
				// }
				// }
				log.Println(paramPosted)
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
					log.Printf("failed to read request body:\n\t%v\n", errR)
					w.Write([]byte("bad request's body"))
					return
				}

				// unmarshall json
				errU := json.Unmarshal(bytesBody, &reqBody)
				if errU != nil {
					log.Printf("failed to unmarshall request body:\n\t%v\n", errU)
					w.Write([]byte("bad request's body"))
					return
				}

				log.Printf("body posted is:\n\t%v", string(bytesBody))

				// check if there is map key(and value) of paramName
				if _, ok := reqBody[*paramName]; !ok {
					errParamNo := fmt.Sprintf("no required param(%s) in body", *paramName)
					log.Println(errParamNo)
					w.Write([]byte(errParamNo))
					return
				}

				// check if param empty
				paramVal = reqBody[*paramName].(string)
				if len(paramVal) == 0 {
					errParamEmpty := fmt.Sprintf("empty param(%s) in body", *paramName)
					log.Println(errParamEmpty)
					w.Write([]byte(errParamEmpty))
					return
				}

				// check body condition
				if len(*bodyCondition) != 0 {
					bodyConditionKey = strings.Split(*bodyCondition, ":")[0]
					bodyConditionVal = strings.Split(*bodyCondition, ":")[1]
				}
				// check only if flag is not empty
				if *bodyCondition != "" {
					if reqBody[bodyConditionKey] != bodyConditionVal {
						log.Printf("additional condition for request body is not met: '%s'", *bodyCondition)
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
					// TODO: add 'error' email
					log.Fatalf("failed to open db:\n\t%v", err)
				}
				defer db.Close()
				// insert name param into db
				postedDate := time.Now().Format("02.01.2006 15:04:05")
				query := fmt.Sprintf("INSERT INTO %s (%s, %s) values('%s', '%s')", dbDataTable, dbValueColumn, dbPostedDateColumn, paramVal, postedDate)
				_, errI := db.Exec(query)

				if errI != nil {
					paramDbInsert := fmt.Sprintf("failed to insert '%s' param into db('%s'):\n\t%v\n", paramVal, dbFile, errI)

					// repeat only if sqlite3 error "sqlite3: unable to open database file"
					regexpErrUnique := regexp.MustCompile("sqlite3: unable to open database file")
					errorStr := fmt.Sprintf("%v", errI)
					if len(regexpErrUnique.Find([]byte(errorStr))) != 0 {
						log.Println(paramDbInsert)
						log.Printf("attemp %d/3 to insert data into db failed, trying again in 5 sec\n", i)
						time.Sleep(5 * time.Second)
						db.Close()
						// stop attempts to insert if it's 3d attempt already
						if i == 3 {
							log.Println("all 3 attempts is failed")
							log.Println(paramDbInsert)
							// mail this error if mailing option is on
							if *mailingOpt {
								mailErr = mailing.SendPlainEmailWoAuth(mailingFile, "error", appName, []byte(paramDbInsert))
								if mailErr != nil {
									log.Printf("failed to send email:\n\t%v", mailErr)
								}
							}
							return
						}
						continue
					}

					// if sqlite3 error not "sqlite3: unable to open database file" - return
					db.Close()
					log.Println(paramDbInsert)
					// mail this Derror if mailing option is on
					if *mailingOpt {
						mailErr = mailing.SendPlainEmailWoAuth(mailingFile, "error", appName, []byte(paramDbInsert))
						if mailErr != nil {
							log.Printf("failed to send email:\n\t%v", mailErr)
						}
					}
					return
				}

				paramProcessed := fmt.Sprintf("%s param successfully processed, waiting for next request", paramVal)
				// mail this if mailing option is on
				if *mailingOpt {
					mailErr = mailing.SendPlainEmailWoAuth(mailingFile, "report", appName, []byte(paramProcessed))
					if mailErr != nil {
						log.Printf("failed to send email:\n\t%v", mailErr)
					}
				}

				log.Println(paramProcessed)
				db.Close()
				break
			}

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
