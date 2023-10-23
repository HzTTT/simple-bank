package main

import (
	"database/sql"
	"log"

	"github.com/HzTTT/simple_bank/api"
	db "github.com/HzTTT/simple_bank/db/sqlc"
	"github.com/HzTTT/simple_bank/util"

	_ "github.com/lib/pq"
)


func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:",err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:",err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:",err)
	}

}