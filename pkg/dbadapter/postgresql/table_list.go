package postgresql

import (
	"fmt"
	"context"
	"database/sql"

	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/auth"
	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/session"
)

type TableRecord struct {
	Name string
	Schema string
	Type string
	Owner string
	
}

type TableList struct {
	*PostgreSQLDBAdapter

	database string
	tables []TableRecord
}

func NewTableList(adapter *PostgreSQLDBAdapter, database string) *TableList {
	return &TableList{
		PostgreSQLDBAdapter: adapter,
		database: database,
	}
}
func (tl *TableList) GetTitle() string {
	return "Tables"
}

func (tl *TableList) GetContent(session *session.Session) tview.Primitive {
	tableTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	tableTable.SetCell(0, 0, tview.NewTableCell("Schema").SetSelectable(false).SetExpansion(1))
	tableTable.SetCell(0, 1, tview.NewTableCell("Name").SetSelectable(false).SetExpansion(1))
	tableTable.SetCell(0, 2, tview.NewTableCell("Type").SetSelectable(false).SetExpansion(1))
	tableTable.SetCell(0, 3, tview.NewTableCell("Owner").SetSelectable(false).SetExpansion(1))
	tableTable.SetCell(1, 0, tview.NewTableCell("Loading tables..."))

	go func() {
		tables, _ := listTables(tl.instance, tl.user, tl.database)
		// session.App.Stop()
		// fmt.Printf("%v", tables)
		for i, table := range tables {
			tableTable.SetCell(i+1, 0, tview.NewTableCell(table.Schema))
			tableTable.SetCell(i+1, 1, tview.NewTableCell(table.Name))
			tableTable.SetCell(i+1, 2, tview.NewTableCell(table.Type))
			tableTable.SetCell(i+1, 3, tview.NewTableCell(table.Owner))

			session.App.Draw()
		}
		
	}()

	return tableTable
}

func listTables(instance *config.InstanceConfig, user *config.UserConfig, database string) ([]TableRecord, error) {
	username := user.Username
	host := instance.Host
	port := instance.Port

	authConfig, err := auth.GetAuth(user.AuthType, user.AuthParams)
	if err != nil {
		return nil, err
	}

	password := authConfig.GetCredential()
	connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", username, password, host, port, database)

	db, err := sql.Open("postgres", connectionUri)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT n.nspname as "Schema",
c.relname as "Name", 
CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 's' THEN 'special' END as "Type",
pg_catalog.pg_get_userbyid(c.relowner) as "Owner"
FROM pg_catalog.pg_class c
    LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind IN ('r','')
    AND n.nspname <> 'pg_catalog'
    AND n.nspname <> 'information_schema'
    AND n.nspname !~ '^pg_toast'
AND pg_catalog.pg_table_is_visible(c.oid)
ORDER BY 1,2;   `)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableRecord
	for rows.Next() {
		var name string
		var schema string
		var tableType string
		var owner string
		if err := rows.Scan(&schema, &name, &tableType, &owner); err != nil {
			return nil, err
		}

		table := TableRecord{
			Name: name,
			Schema: schema,
			Type: tableType,
			Owner: owner,
		}

		tables = append(tables, table)
	}

	//return tables
	// tableRecord := TableRecord{
	// 	Name: "Test",
	// }
	// return []TableRecord{tableRecord}, nil

	return tables, nil
}
