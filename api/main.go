package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"agones-minecraft/cmd/api"
	"agones-minecraft/cmd/migrations"
)

func Run() error {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "api",
				Usage: "start api server",
				Action: func(c *cli.Context) error {
					return api.Run()
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					return migrations.Run()
				},
			},
		},
	}

	return app.Run(os.Args)
}
