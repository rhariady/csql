package main

import (
	"os"

	_ "github.com/lib/pq"
	"github.com/mattn/go-isatty"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/session"
	"github.com/rhariady/csql/pkg/app"
	_ "github.com/rhariady/csql/pkg/dbadapter"
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

		application := tview.NewApplication()
		
		session := session.NewSession(application, cfg)
		instanceList := app.NewInstanceList()
		session.SetView(instanceList)

		// var sessions []*session.Session
		// sessions = make([]*session.Session, 10)

		// sessions = append(sessions, session)


		if err := application.Run(); err != nil {
			panic(err)
		}

		// dbadapter.CloseAllAdapter()
	}
}

