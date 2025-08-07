package app

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/discovery"
	"github.com/rhariady/csql/pkg/session"
)

type InstanceList struct{
	instanceTable *tview.Table
}

func (i *InstanceList) GetTitle() string {
	return "Instances"
}

func (i *InstanceList) GetContent(s *session.Session) tview.Primitive {
	// Create the main database table
	i.instanceTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 1)

	i.RefreshInstanceTable(s)

	// 	Set the selected function for the table (triggered by Enter key)
	i.instanceTable.SetSelectedFunc(func(row int, column int) {
		if row == 0 { // Skip header row
			return
		}
		instanceName := i.instanceTable.GetCell(row, 0).Text
		instance := s.Config.GetInstance(instanceName)
		userList := NewUserList(instance)
		s.ShowModal(userList)
	})

	// Set input capture for 'a' key to trigger the same selection logic
	i.instanceTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'a' {
			// ShowAddDatabasesForm(app.app, app.active_session.pages, databaseInstanceList)
			discoverDatabase := NewDiscoverDatabase(i)
			s.ShowModal(discoverDatabase)
			return nil // Consume the event
		}
		if event.Rune() == 'd' {
			row, column := i.instanceTable.GetSelection()
			instanceName := i.instanceTable.GetCell(row, column).Text
			
			messages := fmt.Sprintf(`Are you sure you want to remove this instance:

%s`, instanceName)

			s.ShowAlert(messages, func(s *session.Session){
				err := s.Config.RemoveInstance(instanceName)
				if err != nil {
					s.ShowMessage(fmt.Sprintf("Error: \n%s", err), true)
				} else {
					i.instanceTable.RemoveRow(row)
					s.ShowMessage(fmt.Sprintf("Instance %s has been removed", instanceName), true)
				}
			}, func(s *session.Session){})

		}
		return event
	})

	return i.instanceTable

}

func AddInstanceForm() {
	fmt.Println("test")
}

func (i *InstanceList) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("[a]", "Add new instance(s)"),
		session.NewKeyBinding("[d]", "Remove instance"),
		session.NewKeyBinding("<enter>", "Connect to instance"),
	}
	return
}

func (i *InstanceList) GetInfo() (info []session.Info) {
	return
}

func (i *InstanceList) ExecuteCommand(s *session.Session, command string) error {
	return nil
}

func NewInstanceList() *InstanceList {
	return &InstanceList{}
}

func (i *InstanceList) RefreshInstanceTable(session *session.Session) {
	i.instanceTable.Clear()
	i.instanceTable.SetCell(0, 0, tview.NewTableCell("Name").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Type").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Host").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 3, tview.NewTableCell("Port").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 4, tview.NewTableCell("Source").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 5, tview.NewTableCell("Params").SetExpansion(2).SetSelectable(false))

	instancesName := slices.Sorted(maps.Keys(session.Config.Instances))

	row := 1
	for _, name := range instancesName {
		instance := session.Config.Instances[name]
		discovery, err := discovery.GetDiscovery(instance.Source)
		var sourceLabel string
		if err != nil {
			sourceLabel = ""
		}
		sourceLabel = discovery.GetLabel()

		var param_list []string
		for param_key, param_value := range instance.Params {
			param_list = append(param_list, fmt.Sprintf("[%s: %s]", param_key, param_value))
		}
		params := strings.Join(param_list, " ")

		i.instanceTable.SetCell(row, 0, tview.NewTableCell(name))
		i.instanceTable.SetCell(row, 1, tview.NewTableCell(instance.Type))
		i.instanceTable.SetCell(row, 2, tview.NewTableCell(instance.Host))
		i.instanceTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprint(instance.Port)))
		i.instanceTable.SetCell(row, 4, tview.NewTableCell(sourceLabel))
		i.instanceTable.SetCell(row, 5, tview.NewTableCell(params))
		row++
	}
}

