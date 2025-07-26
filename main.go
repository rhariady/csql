package main

import (
	"os"

	_ "github.com/lib/pq"
	"github.com/mattn/go-isatty"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/app"
)

func main() {
	config.CheckConfigFile()
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		panic("This application is intended to be run in an interactive terminal.")
	} else {

		app := app.NewApplication(cfg)

		if err := app.Run(); err != nil {
			panic(err)
		}
	}
}

