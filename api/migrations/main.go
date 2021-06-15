package main

import (
	"agones-minecraft/config"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

func main() {
	if err := Migrations.DiscoverCaller(); err != nil {
		log.Fatal(err)
	}

	config.InitConfig()

	dbConfig := config.GetDBConfig()

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Hostname,
		dbConfig.Port,
		dbConfig.Name,
	)

	fmt.Println(dsn)

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	// disable prepared statements
	config.PreferSimpleProtocol = true

	sqldb := stdlib.OpenDB(*config)

	if err != nil {
		log.Fatal("error connecting to datbase")
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	if err := Migrations.Init(context.Background(), db); err != nil {
		log.Fatal(err)
	}

	if err := Migrations.Unlock(context.Background(), db); err != nil {
		log.Fatal(err)
	}

	if err := Migrations.Migrate(context.Background(), db); err != nil {
		log.Fatal(err)
	}
}
