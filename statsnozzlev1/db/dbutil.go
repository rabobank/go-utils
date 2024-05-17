package db

import (
	"bytes"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rabobank/go-utils/statsnozzlev1/conf"
	"github.com/rabobank/go-utils/statsnozzlev1/model"
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
		log.Printf("database %s does not exist, creating it...\n", dbURL.Path)
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

func InsertStats(stats []model.Stats) error {
	if tx, err := Database.Begin(); err != nil {
		return err
	} else {
		buf := bytes.NewBuffer([]byte("INSERT INTO statsnozzle (time,ip,peer_type,method,status_code,content_length,uri,remote, remote_port,forwarded,useragent) VALUES "))
		for i := 0; i < len(stats); i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("(?,?,?,?,?,?,?,?,?,?,?)")
		}
		if bulkInsertStmt, err := tx.Prepare(buf.String()); err != nil {
			return err
		} else {
			var values []interface{}
			for _, stat := range stats {
				values = append(values, stat.Time)
				values = append(values, stat.IP)
				values = append(values, stat.PeerType)
				values = append(values, stat.Method)
				values = append(values, stat.StatusCode)
				values = append(values, stat.ContentLength)
				values = append(values, stat.URI)
				values = append(values, stat.Remote)
				values = append(values, stat.RemotePort)
				values = append(values, stat.ForwardedFor)
				values = append(values, stat.UserAgent)
			}
			if _, err = bulkInsertStmt.Exec(values...); err != nil {
				return err
			}
			values = values[0:0]
			return tx.Commit()
		}
	}
}
