package main

import (
	"database/sql"
	"log"
)

var (
	DATABASE *sql.DB
)

func dbConnect() {
	var e error
	DATABASE, e = sql.Open("sqlite3", ":memory:")
	if e != nil {
		log.Fatal(e)
	}
}

func dbCreateTable() bool {
	_, e := DATABASE.Exec(`
		CREATE TABLE channels (
			guild_id BIGINT NOT NULL,
			channel_id BIGINT NOT NULL,
			currently_using BOOL DEFAULT FALSE,
			currently_playing TEXT DEFAULT NULL,
			primary key (guild_id, channel_id)
		)`)

	if e != nil {
		log.Printf("dbCreateTable error: %v", e)
		return false
	}

	return true
}

func dbExec(query string, parameters ...interface{}) int {
	result, e := DATABASE.Exec(query, parameters...)
	if e != nil {
		log.Printf("dbExec error: %v, query: %v", e, query)
		return -1
	}

	affected, e := result.RowsAffected()
	if e != nil {
		log.Printf("dbExec error: %v, query: %v", e, query)
		return -1
	}

	return int(affected)
}

func dbQuery(query string, parameters ...interface{}) [][]string {
	result := [][]string{}

	rows, e := DATABASE.Query(query, parameters...)
	if e != nil {
		log.Printf("dbQuery error: %v, query: %v", e, query)
		return nil
	}
	defer rows.Close()

	columns, e := rows.Columns()
	if e != nil {
		log.Printf("dbQuery error: %v, query: %v", e, query)
		return nil
	}

	tempBytePtr := make([]interface{}, len(columns))
	tempByte := make([][]byte, len(columns))
	tempString := make([]string, len(columns))
	for i := range tempByte {
		tempBytePtr[i] = &tempByte[i]
	}

	for rows.Next() {
		if e := rows.Scan(tempBytePtr...); e != nil {
			log.Printf("dbQuery error: %v, query: %v", e, query)
			return nil
		}

		for i, rawByte := range tempByte {
			if rawByte == nil {
				tempString[i] = "\\N"
			} else {
				tempString[i] = string(rawByte)
			}
		}

		result = append(result, make([]string, len(columns)))
		copy(result[len(result)-1], tempString)
	}

	return result
}

/*
func dbTx(procedure func(*sql.Tx) bool) bool {
	tx, e := DATABASE.Begin()
	if e != nil {
		log.Printf("dbTx Begin error: %v", e)
		return false
	}

	if procedure(tx) {
		e := tx.Commit()
		if e != nil {
			log.Printf("dbTx Commit error: %v", e)
			return false
		}
		return true
	} else {
		e := tx.Rollback()
		if e != nil {
			log.Printf("dbTx Rollback error: %v", e)
			return false
		}
		return false
	}
}
*/
