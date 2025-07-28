package postgresql

import (
	"fmt"
	"context"
	"database/sql"
	_ "strings"

	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
)

type TableQueryRecord struct {
	Name string
	Schema string
	Type string
	Owner string
	
}

type TableQuery struct {
	*PostgreSQLAdapter

	table string
}

func NewTableQuery(adapter *PostgreSQLAdapter, table string) *TableQuery {
	return &TableQuery{
		PostgreSQLAdapter: adapter,
		table: table,
	}
}
func (tq *TableQuery) GetTitle() string {
	return "Query Result"
}

func (tq *TableQuery) GetContent(session *session.Session) tview.Primitive {
	queryResultTable := tview.NewTable().
		SetBorders(true).
		SetSelectable(false, true)

	go func() {
		// session.App.Stop()
		// columns, _ := getTableColumn(tq.conn, tq.table)
		// session.App.Stop()

		rows, columns, _ := queryTable(tq.conn, tq.table)
		// session.App.Stop()
		// fmt.Printf("%v", rows)
		// fmt.Printf("%v", err)
		for idx, column := range columns {
			queryResultTable.SetCell(0, idx, tview.NewTableCell(column).SetSelectable(false))
		}
		
		for i, row := range rows {
			for j, column := range columns {
				queryResultTable.SetCell(i+1, j, tview.NewTableCell(row[column]))
			}

			session.App.Draw()
		}
		
	}()

	return queryResultTable
}


func getTableColumn(conn *sql.DB, tableName string) ([]string, error) {
	ctx := context.Background()
	params := fmt.Sprintf("public.%s", tableName)
	// fmt.Print(params)
	// fmt.Print("Test2")
	rows, err := conn.QueryContext(ctx, `SELECT attname            AS col
FROM   pg_attribute
WHERE  attrelid = $1::regclass
AND    attnum > 0
AND    NOT attisdropped
ORDER  BY attnum;`, params)

	// fmt.Print("Test1")
	if err != nil {
		// fmt.Print(err)
		return nil, err
	}
	defer rows.Close()

	columns := make([]string, 0)
	for rows.Next() {
		// fmt.Print("Next")
		//var table string
		var column string
		//var datatype string
		if err := rows.Scan(&column); err != nil {
			// fmt.Print(err)
			return nil, err
		}
		// fmt.Print(column)
		columns = append(columns, column)
	}

	return columns, nil
}

func queryTable(conn *sql.DB, tableName string) ([]map[string]string, []string, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 100", tableName)
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
	// return [][]string{[]string{"1", "y"}}
}
