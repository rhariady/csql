package postgresql

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
)

type DatabaseRecord struct {
	Name string
	Owner string
	Encoding string
	Collate string
	Ctype string
	AccessPrivileges string
}

type DatabaseList struct {
	*PostgreSQLAdapter
}

func (d *DatabaseList) GetTitle() string {
	return "Databases"
}

func (d *DatabaseList) GetContent(session *session.Session) tview.Primitive {
	// Table for databases
	databaseTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	databaseTable.SetCell(0, 0, tview.NewTableCell("Name").SetSelectable(false).SetExpansion(1))
	databaseTable.SetCell(0, 1, tview.NewTableCell("Owner").SetSelectable(false))
	databaseTable.SetCell(0, 2, tview.NewTableCell("Encoding").SetSelectable(false))
	databaseTable.SetCell(0, 3, tview.NewTableCell("Collate").SetSelectable(false))
	databaseTable.SetCell(0, 4, tview.NewTableCell("Ctype").SetSelectable(false))
	databaseTable.SetCell(0, 5, tview.NewTableCell("Access Privileges").SetSelectable(false))
	databaseTable.SetCell(1, 0, tview.NewTableCell("Loading databases..."))

	// Get databases
	go func() {
		databases, err := d.PostgreSQLAdapter.listDatabases()
		if err != nil {
			session.ShowMessageAsync(fmt.Sprintf("Error:\n%s", err), true)
			return
		}

		// Populate table
		for i, db := range databases {
			databaseTable.SetCell(i+1, 0, tview.NewTableCell(db.Name))
			databaseTable.SetCell(i+1, 1, tview.NewTableCell(db.Owner))
			databaseTable.SetCell(i+1, 2, tview.NewTableCell(db.Encoding))
			databaseTable.SetCell(i+1, 3, tview.NewTableCell(db.Collate))
			databaseTable.SetCell(i+1, 4, tview.NewTableCell(db.Ctype))
			databaseTable.SetCell(i+1, 5, tview.NewTableCell(db.AccessPrivileges))
		}

		// On selection, go to table
		session.App.Draw()
	}()

	databaseTable.SetSelectedFunc(func(row int, column int) {
		if row == 0 { // Skip header
			return
		}
		d.database = databaseTable.GetCell(row, 0).Text
		tableList := NewTableList(d.PostgreSQLAdapter)
		session.SetView(tableList)
	})

	databaseTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey{
		return d.PostgreSQLAdapter.InputCapture(session, event)
	})

	return databaseTable
}

func (i *DatabaseList) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("<enter>", "List database tables"),
	}

	base_keybinding := i.PostgreSQLAdapter.GetKeyBindings()
	keybindings = append(keybindings, base_keybinding...)

	return
}

func NewDatabaseList(adapter *PostgreSQLAdapter) *DatabaseList {
	return &DatabaseList{
		PostgreSQLAdapter: adapter,
	}
}
