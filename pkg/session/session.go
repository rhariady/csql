package session

import (
	"fmt"

	// "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/config"
)

type Session struct {
	App *tview.Application
	Config *config.Config

	pages *tview.Pages
	headerFlex *tview.Flex
	mainFlex *tview.Flex
	commandBar *tview.InputField
}

type KeyBinding struct {
	hint string
	description string
	// function func()
}

func NewKeyBinding(hint string, description string) *KeyBinding {
	return &KeyBinding{
		hint: hint,
		description: description,
		// function: function,
	}
}

type Info struct {
	key string
	value string
}

func NewInfo(key string, value string) Info {
	return Info{
		key: key,
		value: value,
		// function: function,
	}
}

type View interface {
	GetTitle() string
	GetContent(*Session) tview.Primitive
	GetInfo() []Info
	GetKeyBindings() []*KeyBinding
	ExecuteCommand(*Session, string) error
}

func NewSession(app *tview.Application, config *config.Config) *Session {
	pages := tview.NewPages()

	mainFlex := tview.NewFlex()
	mainFlex.SetBorder(true)

	headerFlex := tview.NewFlex()
	commandBar := tview.NewInputField().SetLabel("/").SetFieldBackgroundColor(tcell.ColorBlack)
	commandBar.SetFieldWidth(0)

	commandFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	commandFlex.AddItem(commandBar, 1, 0, true)
	commandFlex.SetBorder(true)

	outerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(headerFlex, 6, 0, false).
		AddItem(commandFlex, 3, 0, false).
		AddItem(mainFlex, 0, 1, true)

	session := Session{
		App: app,
		Config: config,
		pages: pages,
		headerFlex: headerFlex,
		mainFlex: mainFlex,
		commandBar: commandBar,
	}

	pages.AddPage("main", outerFlex, true, true)

	app.SetRoot(pages, true)

	return &session
}

func (s *Session) setInputCapture() {
	s.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == '/' {
			s.commandBar.SetText("")
			s.App.SetFocus(s.commandBar)
			return nil
		}
		return event
	})
}

func (s *Session) clearInputCapture() {
	s.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
}

func (s *Session) ShowHeader(info_list []Info, keyBindings []*KeyBinding) {
	s.headerFlex.Clear()
	logo := tview.NewTextView().SetText(`
   ___________ ____    __
  / ____/ ___// __ \  / /
 / /    \__ \/ / / / / /
/ /___ ___/ / /_/ / / /___
\____//____/\___\_\/_____/`)

	info_grid := tview.NewGrid().
		SetRows(1, 1, 1, 1, 1, 1).
		SetColumns(12, 0)

	for i, info := range info_list {
		info_grid.AddItem(tview.NewTextView().SetText(fmt.Sprintf("%s", info.key)), i, 0, 1, 1, 0, 0, false)
		info_grid.AddItem(tview.NewTextView().SetText(fmt.Sprintf("%s", info.value)), i, 1, 1, 1, 0, 0, false)
	}

	keyLegend := tview.NewGrid().
		SetRows(1, 1, 1, 1, 1, 1).
		SetColumns(8, 0)

	for i, binding := range keyBindings {
		x := i / 6
		y := i % 6
		keyLegend.AddItem(tview.NewTextView().SetText(fmt.Sprintf("%s", binding.hint)), y, x, 1, 1, 0, 0, false)
		keyLegend.AddItem(tview.NewTextView().SetText(fmt.Sprintf("%s", binding.description)), y, x+1, 1, 1, 0, 0, false)
	}

	// databaseInfo := tview
	s.headerFlex.AddItem(info_grid, 0, 1, false).
		AddItem(keyLegend, 0, 1, true).
		AddItem(logo, 28, 0, false)
}

func (s *Session) SetView(view View) {
	info := view.GetInfo()
	keybindings := view.GetKeyBindings()
	s.ShowHeader(info, keybindings)

	content := view.GetContent(s)
	s.mainFlex.SetTitle(view.GetTitle())
	s.mainFlex.Clear()
	s.mainFlex.AddItem(content, 0, 1, true)

	s.commandBar.
		SetFinishedFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				command := s.commandBar.GetText()
				s.commandBar.SetText("")
				view.ExecuteCommand(s, command)
			case tcell.KeyEsc:
				s.commandBar.SetText("")
				s.App.SetFocus(s.mainFlex.GetItem(0))
			}
		})

	s.App.SetFocus(content)
	s.setInputCapture()

	// keybindingMap := make(map[string]*KeyBinding)
	// for _, keybinding := range keybindings {
	// 	keybindingMap[keybinding.key] = keybinding
	// }

	// content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	rune := event.Rune()
	// 	keybinding, ok := keybindingMap[rune]
	// 	if ok {
	// 		return nil
	// 	}

	// 	return event
	// })
}

func (s *Session) ShowModal(view View) {
	s.clearInputCapture()

	content := view.GetContent(s)
	keybindings := view.GetKeyBindings()
	keybindings = append(keybindings, NewKeyBinding("<esc>", "Close"))
	modalFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(content, 0, 1, true)

	for _, keybinding := range keybindings {
		legend_text := fmt.Sprintf("%s %s", keybinding.hint, keybinding.description)
		legend := tview.NewTextView().SetText(legend_text).SetTextAlign(tview.AlignCenter).SetWrap(true).SetWordWrap(true)
		modalFlex.AddItem(legend, 1, 0, false)
	}

	modalFlex.SetBorder(true).SetTitle(view.GetTitle())

	rowFlex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modalFlex, 0, 2, true).
			AddItem(nil, 0, 1, false)


	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(rowFlex, 0, 1, true).
		AddItem(nil, 0, 1, false)

	modalFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey{
		if event.Key() == tcell.KeyEsc {
			s.CloseModal()
			return nil
		}
		return event
	})
	s.pages.AddPage("modal", modal, true, true)

	s.App.SetFocus(content)
}

func (s *Session) CloseModal() {
	s.pages.RemovePage("modal")
	s.setInputCapture()
}

func (s *Session) ShowMessageAsync(text string, wait bool) {
	s.App.QueueUpdateDraw(func() {
		s.ShowMessage(text, wait)
	})
}

func (s *Session) ShowMessage(text string, wait bool) {
	modal := tview.NewModal().
		SetText(text)

	if wait {
		modal.
			AddButtons([]string{"OK"}).SetDoneFunc(func(index int, label string){
			s.CloseMessage()
		})
	}
	s.pages.AddPage("message", modal, true, true)
}

func (s *Session) CloseMessageAsync() {
	s.App.QueueUpdateDraw(func() {
		s.CloseMessage()
	})
}

func (s *Session) CloseMessage() {
	s.pages.RemovePage("message")
}

func (s *Session) ShowAlertAsync(text string, ok func(*Session), cancel func(*Session)) {
	s.App.QueueUpdateDraw(func() {
		s.ShowAlert(text, ok, cancel)
	})
}

func (s *Session) ShowAlert(text string, ok func(*Session), cancel func(*Session)) {
	modal := tview.NewModal().SetText(text)

	var buttons []string

	if ok != nil {
		buttons = append(buttons, "OK")
	}

	if cancel != nil {
		buttons = append(buttons, "Cancel")
	}

	if len(buttons) > 0 {
		modal.AddButtons(buttons)
	}

	modal.SetDoneFunc(func(index int, label string) {
		s.pages.RemovePage("alert")
		switch label {
		case "OK":
			ok(s)
		case "Cancel":
			cancel(s)
		}
	})

	s.pages.AddPage("alert", modal, true, true)
}

func (s *Session) CloseAlertAsync() {
	s.App.QueueUpdateDraw(func() {
		s.CloseAlert()
	})
}

func (s *Session) CloseAlert() {
	s.pages.RemovePage("alert")
}
