package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
)

type InstanceList struct{
	databaseInstanceTable *tview.Table
}

func (i *InstanceList) GetTitle() string {
	return "Instances"
}

func (i *InstanceList) GetContent(session *session.Session) tview.Primitive {
	// Create the main database table
	i.databaseInstanceTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	i.RefreshInstanceTable(session)

	// 	Set the selected function for the table (triggered by Enter key)
	i.databaseInstanceTable.SetSelectedFunc(func(row int, column int) {
		if row == 0 { // Skip header row
			return
		}
		instanceName := i.databaseInstanceTable.GetCell(row, 0).Text
		instance := session.Config.GetInstance(instanceName)
		userList := NewUserList(instance)
		session.ShowModal(userList)
	})

	// Set input capture for 'a' key to trigger the same selection logic
	i.databaseInstanceTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'a' {
			// ShowAddDatabasesForm(app.app, app.active_session.pages, databaseInstanceList)
			discoverDatabase := NewDiscoverDatabase(i)
			session.ShowModal(discoverDatabase)
			return nil // Consume the event
		}
		return event
	})

	return i.databaseInstanceTable

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

func (i *InstanceList) RefreshInstanceTable(session *session.Session) {
	i.databaseInstanceTable.Clear()
	i.databaseInstanceTable.SetCell(0, 0, tview.NewTableCell("Name").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Type").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Host").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 3, tview.NewTableCell("Port").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 4, tview.NewTableCell("Params").SetExpansion(1).SetSelectable(false))

	i.databaseInstanceTable.SetWrapSelection(true, true)
	row := 1
	for name, instance := range session.Config.Instances {
		i.databaseInstanceTable.SetCell(row, 0, tview.NewTableCell(name))
		i.databaseInstanceTable.SetCell(row, 1, tview.NewTableCell(instance.Type))
		i.databaseInstanceTable.SetCell(row, 2, tview.NewTableCell(instance.Host))
		i.databaseInstanceTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprint(instance.Port)))
		i.databaseInstanceTable.SetCell(row, 4, tview.NewTableCell(fmt.Sprint(instance.Params)))
		row++
	}
}

