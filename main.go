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

		application := app.NewApplication(cfg)

		var sessions []*app.Session
		sessions = make([]*app.Session, 10)
		
		sessions = append(sessions, app.NewSession(application))

		if err := application.Start(); err != nil {
			panic(err)
		}
	}
}

