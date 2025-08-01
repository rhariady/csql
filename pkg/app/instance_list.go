package app

import (
	"fmt"
	
	_ "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
)

type InstanceList struct{}

func (i *InstanceList) GetTitle() string {
	return "Instances"
}

func (i *InstanceList) GetContent(session *session.Session) tview.Primitive {
	// Create the main database table
	databaseInstanceList := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	RefreshInstanceTable(session, databaseInstanceList)

	// 	Set the selected function for the table (triggered by Enter key)
	databaseInstanceList.SetSelectedFunc(func(row int, column int) {
		if row == 0 { // Skip header row
			return
		}
		instanceName := databaseInstanceList.GetCell(row, 0).Text
		instance := session.Config.GetInstance(instanceName)
		userList := NewUserList(instance)
		session.ShowModal(userList)
	})

	// Set input capture for 'a' key to trigger the same selection logic
	// databaseInstanceList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	if event.Rune() == 'a' {
	// 		ShowAddDatabasesForm(app.app, app.active_session.pages, databaseInstanceList)
	// 		return nil // Consume the event
	// 	}
	// 	return event
	// })

	return databaseInstanceList

}

func AddInstanceForm() {
	fmt.Println("test")
}

func (i *InstanceList) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("[a]", "Add new instance"),
		session.NewKeyBinding("<enter>", "Select instance"),
	}
	return
}

func NewInstanceList() *InstanceList {
	return &InstanceList{}
}

func RefreshInstanceTable(session *session.Session, table *tview.Table) {
	table.Clear()
	table.SetCell(0, 0, tview.NewTableCell("Name").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Type").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Host").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 3, tview.NewTableCell("Port").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 4, tview.NewTableCell("Params").SetExpansion(1).SetSelectable(false))

	table.SetWrapSelection(true, true)
	row := 1
	for name, instance := range session.Config.Instances {
		table.SetCell(row, 0, tview.NewTableCell(name))
		table.SetCell(row, 1, tview.NewTableCell(instance.Type))
		table.SetCell(row, 2, tview.NewTableCell(instance.Host))
		table.SetCell(row, 3, tview.NewTableCell(fmt.Sprint(instance.Port)))
		table.SetCell(row, 4, tview.NewTableCell(fmt.Sprint(instance.Params)))
		row++
	}
}

