package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shivangp0208/bank_application/api"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/util"
)

var conn *sql.DB
var err error
var config util.Config
var logger = util.GetLogger()

func init() {
	config = util.GetConfig()
	conn, err = sql.Open(config.DBDriver, config.DBSource)
	logger.Printf("init main: dbDriver: %s and dbSource: %s", config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("unable to open db connection: %v", err)
	}
	logger.Printf("successfull opening up the db connection")
}

func main() {

	store := db.NewStore(conn)
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatalf("unable to create new server due to err %v", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatalf("unable to start the server with address %s due to err %v", config.ServerAddress, err)
	}
}
