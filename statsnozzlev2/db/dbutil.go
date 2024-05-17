package db

import (
	"bytes"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rabobank/go-utils/statsnozzlev2/conf"
	"log"
	"net/url"
	"os"
)

var Database *sql.DB

func InitDB() {
	var DbExists bool
	var err error
	dbURL, err := url.Parse("file:" + conf.DBFile)
	if err != nil {
		log.Fatalf("failed parsing database url %s, error: %s", dbURL, err.Error())
	}
	if _, err = os.Stat(dbURL.Opaque); err != nil && dbURL.Scheme == "file" {
		log.Printf("database %s does not exist, creating it...\n", dbURL)
		DbExists = false
	} else {
		DbExists = true
	}

	Database, err = sql.Open("sqlite3", "file:"+conf.DBFile)
	if err != nil {
		log.Fatal(err)
	}

	// a simple but effective way to not get "database is locked" from sqlite3
	Database.SetMaxOpenConns(1)

	if !DbExists {
		var sqlStmts []byte
		if sqlStmts, err = os.ReadFile(conf.CreateTablesFile); err != nil {
			log.Fatal(err)
		} else {
			if _, err = Database.Exec(string(sqlStmts)); err != nil {
				log.Fatalf("%q: %s\n", err, sqlStmts)
			}
		}
	}
}

func InsertStats(payloadBatch [][]string) error {
	if tx, err := Database.Begin(); err != nil {
		return err
	} else {
		buf := bytes.NewBuffer([]byte("INSERT INTO statsnozzle (route,time,method,path,protocol,response_code,body_size,response_size,referrer,user_agent,remote_addr,upstream_addr,x_forwarded_for,x_forwarded_proto,vcap_request_id,response_time,gorouter_time,app_id,app_index,instance_id,x_cf_router_error,x_rabo_client_ip,x_session_id,traceparent,tracestate,org_name,space_name,app_name) VALUES "))
		for i := 0; i < len(payloadBatch); i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		}
		if bulkInsertStmt, err := tx.Prepare(buf.String()); err != nil {
			return err
		} else {
			var values []interface{}
			for _, payloadSlice := range payloadBatch {
				for i := 0; i < len(payloadSlice); i++ {
					values = append(values, payloadSlice[i])
				}
			}
			if _, err = bulkInsertStmt.Exec(values...); err != nil {
				return err
			}
			values = values[0:0]
			return tx.Commit()
		}
	}
	return nil
}
