
package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	_ "strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
)

type QueryEditor struct {
	*PostgreSQLAdapter
	query string
}

func NewQueryEditor(adapter *PostgreSQLAdapter, query string) *QueryEditor {
	return &QueryEditor{
		PostgreSQLAdapter: adapter,
		query: query,
	}
}

func (tq *QueryEditor) GetTitle() string {
	return "Query"
}

func (tq *QueryEditor) GetContent(session *session.Session) tview.Primitive {
	queryInput := tview.NewTextArea()
	queryInput.SetText(tq.query, true)

	queryResultTable := tview.NewTable().
		SetBorders(true).
		SetSelectable(false, false).
		SetFixed(1, 0)

	queryInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlX {
			tq.query = queryInput.GetText()
			go func() {
				rows, columns, err := executeQuery(tq.conn, tq.query)
				if err != nil {
					session.ShowMessageAsync(fmt.Sprintf("Error: %s", err), true)
					return
				}

				session.App.QueueUpdateDraw(func() {
					queryResultTable.Clear()
					for idx, column := range columns {
						queryResultTable.SetCell(0, idx, tview.NewTableCell(column).SetSelectable(false))
					}

					for i, row := range rows {
						for j, column := range columns {
							queryResultTable.SetCell(i+1, j, tview.NewTableCell(row[column]))
						}
					}
					queryResultTable.ScrollToBeginning()
				})
			}()
			return nil
		}
		return event
	})

	queryInput.SetBorder(true)
	queryResultTable.SetBorder(true)
	
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(queryInput, 0, 1, true).
		AddItem(queryResultTable, 0, 3, false)

	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if queryInput.HasFocus() {
				session.App.SetFocus(queryResultTable)
			} else {
				session.App.SetFocus(queryInput)
			}
			return nil
		}

		if event.Key() == tcell.KeyEsc {
			tableList := NewTableList(tq.PostgreSQLAdapter)
			session.SetView(tableList)
		}
		return event
	})

	return layout
}

func (i *QueryEditor) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("<ctrl-x>", "Execute query"),
		session.NewKeyBinding("<tab>", "Switch focus"),
		session.NewKeyBinding("<esc>", "Go back to table list"),
	}

	return
}

func executeQuery(conn *sql.DB, query string) ([]map[string]string, []string, error) {
	ctx := context.Background()

	rows, err := conn.QueryContext(ctx, query)

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var result []map[string]string
	columns, _ := rows.Columns()
	for rows.Next() {
		x := make([]interface{}, len(columns))
		scans := make([]string, len(columns))
		row := make(map[string]string)

		for i := range scans {
			x[i] = &scans[i]
		}
		rows.Scan(x...)

		for i, v := range scans {
			row[columns[i]] = fmt.Sprintf("%s", v)
		}
		result = append(result, row)
	}

	return result, columns, nil
}

func (i *QueryEditor) GetInfo() (info []session.Info) {
	info = i.PostgreSQLAdapter.GetInfo()
	return
}
