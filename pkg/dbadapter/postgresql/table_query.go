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

type TableQueryRecord struct {
	Name   string
	Schema string
	Type   string
	Owner  string
}

type TableQuery struct {
	*PostgreSQLAdapter

	table string
}

func NewTableQuery(adapter *PostgreSQLAdapter, table string) *TableQuery {
	return &TableQuery{
		PostgreSQLAdapter: adapter,
		table:             table,
	}
}

func (tq *TableQuery) GetTitle() string {
	return "Query Result"
}

func (tq *TableQuery) GetContent(session *session.Session) tview.Primitive {
	queryResultTable := tview.NewTable().
		SetBorders(true).
		SetSelectable(false, false).
		SetFixed(1, 0)

	go func() {
		rows, columns, err := queryTable(tq.conn, tq.table)
		if err != nil {
			session.ShowMessageAsync(fmt.Sprintf("Error: %s", err), true)
		}
		for idx, column := range columns {
			queryResultTable.SetCell(0, idx, tview.NewTableCell(column).SetSelectable(false))
		}

		for i, row := range rows {
			for j, column := range columns {
				queryResultTable.SetCell(i+1, j, tview.NewTableCell(row[column]))
			}

			queryResultTable.ScrollToBeginning()
			session.App.Draw()
		}

	}()

	queryResultTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc {
			tableList := NewTableList(tq.PostgreSQLAdapter)
			session.SetView(tableList)
		}
	})

	queryResultTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return tq.InputCapture(session, event)
	})

	return queryResultTable
}

func (i *TableQuery) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("<escape>", "Go back to table list"),
	}

	base_keybinding := i.PostgreSQLAdapter.GetKeyBindings()
	keybindings = append(keybindings, base_keybinding...)

	return
}

func queryTable(conn *sql.DB, tableName string) (results []map[string]string, columns []string, err error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 100", tableName)
	ctx := context.Background()

	rows, err := conn.QueryContext(ctx, query)

	defer func() {
		r_err := rows.Close()
		if r_err != nil {
			err = r_err
		}
	}()

	if err != nil {
		return nil, nil, err
	}

	var result []map[string]string
	columns, err = rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	for rows.Next() {
		x := make([]interface{}, len(columns))
		scans := make([]string, len(columns))
		row := make(map[string]string)

		for i := range scans {
			x[i] = &scans[i]
		}
		err = rows.Scan(x...)
		if err != nil {
			return nil, nil, err
		}

		for i, v := range scans {
			row[columns[i]] = v
		}
		result = append(result, row)
	}

	return result, columns, nil
	// return [][]string{[]string{"1", "y"}}
}
