package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dbDriver = "mysql"
	dbSource = "shivang:12345678@tcp(localhost:3306)/bank_application?parseTime=true"
)

var testQueries *Queries
var testDB *sql.DB
var err error

func init() {
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("unable to open db connection due to %v", err)
	}
}

func TestMain(m *testing.M) {
	// we are able to give the conn easily to this because sql.DB struct implements the DBTX interface created by sqlc tool
	testQueries = New(testDB)
	os.Exit(m.Run())
}
