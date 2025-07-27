package app

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/config"
)

var cfg *config.Config

type Application struct {
	app *tview.Application
	config *config.Config
}

func (a *Application) Start() error {
	err := a.app.Run()
	return err
}

func NewApplication(config *config.Config) *Application {
	cfg = config
	app := tview.NewApplication()

	application := Application{
		app: app,
		config: config,
	}

	return &application
}

func ShowHeader(app *tview.Application, pages *tview.Pages, flex *tview.Flex) {
	logo := tview.NewTextView().SetText(`
   ___________ ____    __
  / ____/ ___// __ \  / /
 / /    \__ \/ / / / / /
/ /___ ___/ / /_/ / / /___
\____//____/\___\_\/_____/`)

	keyBindings := []struct {
		Key     string
		Purpose string
	}{
		{"q", "quit"},
		{"a", "add instance"},
		{"enter", "select"},
	}

	keyLegend := tview.NewGrid().
		SetRows(1, 1, 1, 1, 1, 1)

	for i, binding := range keyBindings {
		x := i / 6
		y := i % 6
		keyLegend.AddItem(tview.NewTextView().SetText(fmt.Sprintf("<%s> %s", binding.Key, binding.Purpose)), y, x, 1, 1, 0, 0, false)
	}

	flex.AddItem(keyLegend, 0, 1, false).
		AddItem(tview.NewBox(), 0, 2, true).
		AddItem(logo, 28, 0, false)
}

