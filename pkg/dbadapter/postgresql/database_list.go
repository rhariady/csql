package postgresql

import (
	"fmt"
	"context"
	"database/sql"
	
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/auth"
	"github.com/rhariady/csql/pkg/session"
	"github.com/rhariady/csql/pkg/config"
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
	*PostgreSQLDBAdapter

	// instance *config.InstanceConfig
	// user *config.UserConfig
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
		databases, _ := listDatabases(d.instance, d.user)
		// if err != nil {
		// 	// Show an error modal
		// 	errorModal := tview.NewModal().
		// 		SetText(fmt.Sprintf("Error loading databases: %v", err)).
		// 		AddButtons([]string{"OK"}).
		// 		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		// 			session.pages.RemovePage("errorModal")
		// 			// session.app.app.SetFocus(userTable)
		// 		})
		// 	session.pages.AddPage("errorModal", errorModal, true, true)
		// 	session.app.SetFocus(errorModal)
		// }

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
		dbName := databaseTable.GetCell(row, 0).Text
		tableList := NewTableList(d.PostgreSQLDBAdapter, dbName)
		session.SetView(tableList)
		// session.App.Stop() // Stop the tview app to hand over to psql

		// instance := cfg.Instances[d.instanceName]
		// user, _ := instance.GetUserConfig(d.userName)
		// dbAdapter, _ := dbadapter.GetDBAdapter(instance.Type)
		// dbAdapter.RunShell(&instance, user, dbName)

	})

	// Go back on escape
	// databaseTable.SetDoneFunc(func(key tcell.Key) {
	// 	if key == tcell.KeyEscape {
	// 		pages.SwitchToPage("mainTable")
	// 	}
	// })
	return databaseTable
}

// func NewDatabaseList(instance *config.InstanceConfig, user *config.UserConfig) *DatabaseList {
// 	return &DatabaseList{
// 		instance: instance,
// 		user: user,
// 	}
// }

func NewDatabaseList(adapter *PostgreSQLDBAdapter) *DatabaseList {
	return &DatabaseList{
		PostgreSQLDBAdapter: adapter,
	}
}

func listDatabases(instance *config.InstanceConfig, userConfig *config.UserConfig) ([]DatabaseRecord, error) {
	user := userConfig.Username
	host := instance.Host
	port := instance.Port
	dbname := "postgres" // Connect to a default database to list others

	authConfig, err := auth.GetAuth(userConfig.AuthType, userConfig.AuthParams)
	if err != nil {
		return nil, err
	}

	password := authConfig.GetCredential()
	connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)

	db, err := sql.Open("postgres", connectionUri)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// rows, err := db.QueryContext(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false;")
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT
  d.datname AS "Name",
  pg_catalog.pg_get_userbyid(d.datdba) AS "Owner",
  pg_catalog.pg_encoding_to_char(d.encoding) AS "Encoding",
  d.datcollate AS "Collate",
  d.datctype AS "Ctype",
  pg_catalog.array_to_string(d.datacl, E'\n') AS "Access privileges"
FROM
  pg_catalog.pg_database d
WHERE
  datistemplate = false
ORDER BY
  d.datname;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []DatabaseRecord
	for rows.Next() {
		var name string
		var owner string
		var encoding string
		var collate string
		var ctype string
		var accessPrivileges sql.NullString
		if err := rows.Scan(&name, &owner, &encoding, &collate, &ctype, &accessPrivileges); err != nil {
			return nil, err
		}

		database := DatabaseRecord{
			Name: name,
			Owner: owner,
			Encoding: encoding,
			Collate: collate,
			Ctype: ctype,
		}

		if accessPrivileges.Valid {
			database.AccessPrivileges = accessPrivileges.String
		}
		databases = append(databases, database)
	}

	return databases, nil
}
	
