package postgresql

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
)

type TableRecord struct {
	Name string
	Schema string
	Type string
	Owner string
	
}

type TableList struct {
	*PostgreSQLAdapter

	tables []TableRecord
}

func NewTableList(adapter *PostgreSQLAdapter) *TableList {
	return &TableList{
		PostgreSQLAdapter: adapter,
		// database: database,
	}
}
func (tl *TableList) GetTitle() string {
	return "Tables"
}

func (tl *TableList) GetContent(session *session.Session) tview.Primitive {
	tableTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	go func() {
		session.ShowMessageAsync("Loading tables", false)
		
		tables, err := tl.PostgreSQLAdapter.listTables()
		session.CloseMessageAsync()

		if err != nil {
			session.ShowMessageAsync(fmt.Sprintf("Error:\n%s", err), true)
			fmt.Println(err)
		}

		session.App.QueueUpdateDraw(func(){
			tableTable.SetCell(0, 0, tview.NewTableCell("Schema").SetSelectable(false).SetExpansion(1))
			tableTable.SetCell(0, 1, tview.NewTableCell("Name").SetSelectable(false).SetExpansion(1))
			tableTable.SetCell(0, 2, tview.NewTableCell("Type").SetSelectable(false).SetExpansion(1))
			tableTable.SetCell(0, 3, tview.NewTableCell("Owner").SetSelectable(false).SetExpansion(1))

			for i, table := range tables {
				tableTable.SetCell(i+1, 0, tview.NewTableCell(table.Schema))
				tableTable.SetCell(i+1, 1, tview.NewTableCell(table.Name))
				tableTable.SetCell(i+1, 2, tview.NewTableCell(table.Type))
				tableTable.SetCell(i+1, 3, tview.NewTableCell(table.Owner))

			}
		})
		
	}()

	tableTable.SetSelectedFunc(func(row int, column int) {
		if row == 0 { // Skip header
			return
		}
		tableName := tableTable.GetCell(row, 1).Text
		tableQuery := NewTableQuery(tl.PostgreSQLAdapter, tableName)
		session.SetView(tableQuery)
	})

	tableTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey{
		rune := event.Rune()
		switch rune {
		case 'd':
			database_list_modal := NewChangeDatabaseModal(tl.PostgreSQLAdapter, tl)
			session.ShowModal(database_list_modal)
			return nil
		case 'w':
			row, _ := tableTable.GetSelection()
			tableName := tableTable.GetCell(row, 1).Text
			viewQuery := NewQueryEditor(tl.PostgreSQLAdapter, "SELECT * FROM " + tableName)
			session.SetView(viewQuery)
			return nil
		case 's':
			psqlView := NewPsqlView(tl.PostgreSQLAdapter)
			session.SetView(psqlView)
			return nil
		}

		return event
	})
	return tableTable
}

func (i *TableList) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("<enter>", "Query table"),
		session.NewKeyBinding("[w]", "Write query"),
		session.NewKeyBinding("[p]", "Open psql shell"),
	}

	base_keybinding := i.PostgreSQLAdapter.GetKeyBindings()
	keybindings = append(keybindings, base_keybinding...)

	return
}

func (i *TableList) GetInfo() (info []session.Info) {
	info = i.PostgreSQLAdapter.GetInfo()
	return
}

