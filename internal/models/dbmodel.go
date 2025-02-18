package models

import (
	"database/sql"
	"fmt"
	"time"
)

type DbModel struct {
	DB *sql.DB
}

// insert new request data to the db
func (model *DbModel) InsertProcessed(dbFile, table, valCol, postedDateCol, valToInsert string) error {

	// insert name param into db
	postedDate := time.Now().Format("02.01.2006 15:04:05")
	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) values('%s', '%s')",
		table,
		valCol,
		postedDateCol,
		valToInsert,
		postedDate,
	)

	// execute query
	_, errI := model.DB.Exec(query)
	if errI != nil {
		return fmt.Errorf("failed to insert '%s' param into db('%s'):\n\t%v", valToInsert, dbFile, errI)
	}

	return nil
}
