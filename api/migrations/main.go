package main

import (
	"flag"
	"log"
	"net"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"

	"agones-minecraft/config"
)

func main() {
	config.InitConfig()
	dbConfig := config.GetDBConfig()

	db := pg.Connect(&pg.Options{
		Addr:     net.JoinHostPort(dbConfig.Hostname, dbConfig.Port),
		User:     dbConfig.User,
		Password: dbConfig.Password,
		Database: dbConfig.Name,
	})

	flag.Parse()

	old, new, err := migrations.Run(db, flag.Args()...)

	if err != nil {
		log.Fatal(err)
	}

	if new != old {
		log.Printf("migrated from version %d to %d\n", old, new)
	} else {
		log.Printf("version is %d\n", old)
	}
}
