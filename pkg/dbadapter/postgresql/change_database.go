package postgresql

import (
	"fmt"
	
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
)

type ChangeDatabaseModal struct {
	*PostgreSQLAdapter	
}

func NewChangeDatabaseModal(adapter *PostgreSQLAdapter) *ChangeDatabaseModal {
	return &ChangeDatabaseModal{
		PostgreSQLAdapter: adapter,
	}
}

func (d *ChangeDatabaseModal) GetTitle() string {
	return "Select a database"
}

func (d *ChangeDatabaseModal) GetContent(session *session.Session) tview.Primitive {
	databaseTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	databases, err := d.PostgreSQLAdapter.listDatabases()
	if err != nil {
		session.ShowMessage(fmt.Sprintf("Error:\n%s", err), true)
		return nil
	}

	row := 0
	for _, database := range databases {
		databaseTable.SetCell(row, 0, tview.NewTableCell(database.Name).SetExpansion(1))
		row++
	}

	databaseTable.Select(0, 0)

	databaseTable.SetSelectedFunc(func(row int, column int) {
		// databaseList := NewDatabaseList(i.instanceName, userName)
		session.CloseModal()

		newDatabase := databaseTable.GetCell(row, 0).Text

		d.database = newDatabase
		err := d.openConnection()

		if err != nil {
			session.ShowMessage(fmt.Sprintf("Error:\n%s", err), true)
		}

		tableList := NewTableList(d.PostgreSQLAdapter)
		session.SetView(tableList)		
	})

	return databaseTable
	
}

func (d *ChangeDatabaseModal) GetKeyBindings() (keybinding []*session.KeyBinding) {
	return
}

func (d *ChangeDatabaseModal) GetInfo() (infe []session.Info) {
	return
}
