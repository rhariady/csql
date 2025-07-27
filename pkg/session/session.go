package session

import (
	"fmt"
	
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/config"
)

type Session struct {
	App *tview.Application
	Config *config.Config
	
	pages *tview.Pages
	headerFlex *tview.Flex
	mainFlex *tview.Flex
}

type KeyBinding struct {
	Key string
	Purpose string
}

type View interface {
	GetTitle() string
	GetContent(*Session) tview.Primitive
	// 	GetKeyBindings() []KeyBinding
}

func NewSession(app *tview.Application, config *config.Config) *Session {
	pages := tview.NewPages()

	mainFlex := tview.NewFlex()
	mainFlex.SetBorder(true)

	headerFlex := tview.NewFlex()

	outerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(headerFlex, 6, 0, false).
		AddItem(mainFlex, 0, 1, true)

	session := Session{
		App: app,
		Config: config,
		pages: pages,
		headerFlex: headerFlex,
		mainFlex: mainFlex,
	}

	pages.AddPage("main", outerFlex, true, true)

	app.SetRoot(pages, true)

	return &session
}

func (s *Session) ShowHeader() {
	s.headerFlex.Clear()
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

	s.headerFlex.AddItem(keyLegend, 0, 1, false).
		AddItem(tview.NewBox(), 0, 2, true).
		AddItem(logo, 28, 0, false)
}

func (s *Session) SetView(view View) {
	s.ShowHeader()

	content := view.GetContent(s)
	s.mainFlex.SetTitle(view.GetTitle())
	s.mainFlex.Clear()
	s.mainFlex.AddItem(content, 0, 1, true)

	s.App.SetFocus(content)
}

func (s *Session) ShowModal(view View) {
	content := view.GetContent(s)
	modalFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(content, 0, 1, true).
		AddItem(tview.NewTextView().SetText("Press Esc to go back").SetTextAlign(tview.AlignCenter), 1, 1, false)
	modalFlex.SetBorder(true).SetTitle(view.GetTitle())
	
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modalFlex, 0, 1, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)
		
	s.pages.AddPage("modal", modal, true, true)

	s.App.SetFocus(content)
}

func (s *Session) CloseModal() {
	s.pages.RemovePage("modal")
}
