package handleweb

import (
	"log"
	"net/http"

	// sqllite support
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
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
			name := r.URL.Query().Get("name")

			if len(name) == 0 {
				log.Println("empty 'name' param posted")
				w.Write([]byte("Empty name param"))
				return
			}

			// TODO: add check for name regexp, must be(?) "RP\d+"
			log.Printf("Name posted: %s", name)
			w.Write([]byte("OK"))

			// open db
			// TODO: data.db to const/flag
			db, err := sql.Open("sqlite3", "file:data.db")
			if err != nil {
				log.Fatalf("failed to open db:\n\t%v", err)
			}
			defer db.Close()

			// check if name already inserted
			var checkIfAlreadyInserted string
			errS := db.QueryRow("SELECT * FROM requests WHERE name=?", name).Scan(&checkIfAlreadyInserted)
			switch {
			// if name is unique(no result on query then insert)
			case errS == sql.ErrNoRows:
				// insert name param into db
				_, errI := db.Exec("INSERT INTO requests(name) values(?)", name)
				if errI != nil {
					log.Fatalf("failed to insert 'name' param into db:\n\t%v", errI)
				}

				log.Printf("%s param successfully insterted into db, waiting for next request", name)

				db.Close()
				return
			case err != nil:
				log.Fatal(err)
			default:
				log.Printf("%s name is not unique, skipping it", name)
				return
			}
		}

		log.Printf("No 'name' param in POST")
		w.Write([]byte("No 'name' parameter!\n"))
		return
	}

	w.Write([]byte("Only POST allowed!\n"))
	log.Printf("wrong parameter in POST: %v\n", r.URL.String())
}

func registerHanlers() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/api", postHandler)
}

func StartWebServer(address string, mux *http.ServeMux) error {
	registerHanlers()

	if err := http.ListenAndServe(address, mux); err != nil {
		return err
	}

	log.Println("STARTED!")

	return nil
}
