package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/rhariady/csql/pkg/auth"
	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/session"
)

type PostgreSQLAdapter struct {
	// session *session.Session
	instance *config.InstanceConfig
	user     *config.UserConfig
	database string
	conn     *sql.DB
}

func (a *PostgreSQLAdapter) openConnection() error {
	authConfig, err := auth.GetAuth(a.user.AuthType, a.user.AuthParams)
	if err != nil {
		return err
	}

	password, err := authConfig.GetCredential()

	if err != nil {
		return err
	}

	connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", a.user.Username, password, a.instance.Host, a.instance.Port, a.database)

	a.conn, err = sql.Open("postgres", connectionUri)

	if err != nil {
		return err
	}

	// Confirm a successful connection.
	if err := a.conn.Ping(); err != nil {
		return err
	}

	return nil
}

func (a *PostgreSQLAdapter) Connect(session *session.Session, instance *config.InstanceConfig, user *config.UserConfig, database string) error {
	// a.session = session
	a.instance = instance
	a.user = user
	a.database = database

	err := a.openConnection()

	if err != nil {
		return err
	}

	// databaseList := NewDatabaseList(instance, user)
	tableList := NewTableList(a)
	session.SetView(tableList)

	return nil
}

func (a *PostgreSQLAdapter) Close() (err error) {
	if a.conn != nil {
		err = a.conn.Close()
	}
	return
}

func (a *PostgreSQLAdapter) InputCapture(session *session.Session, event *tcell.EventKey) *tcell.EventKey {
	rune := event.Rune()
	switch rune {
	case 'q':
		viewQuery := NewQueryEditor(a, "SELECT * FROM ")
		session.SetView(viewQuery)
		return nil
	case 'd':
		database_list_modal := NewChangeDatabaseModal(a)
		session.ShowModal(database_list_modal)
		return nil
	case 's':
		shellView := NewShellView(a)
		session.SetView(shellView)
		return nil
	}
	return event
}

func (a *PostgreSQLAdapter) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("[q]", "Query Editor"),
		session.NewKeyBinding("[s]", "Start psql shell"),
		session.NewKeyBinding("[d]", "Change database"),
	}
	return
}

func (a *PostgreSQLAdapter) GetInfo() (info []session.Info) {
	info = []session.Info{
		session.NewInfo("Instance", a.instance.Name),
		session.NewInfo("User", a.user.Username),
	}

	if a.database != "" {
		info = append(info, session.NewInfo("Database", a.database))
	}

	return
}

func (a *PostgreSQLAdapter) ExecuteCommand(s *session.Session, command string) error {
	switch command {
	case "table":
		tableList := NewTableList(a)
		s.SetView(tableList)
	case "role":
		roleList := NewRoleList(a)
		s.SetView(roleList)
	case "database":
		databaseList := NewDatabaseList(a)
		s.SetView(databaseList)
	}

	return nil
}

type RoleRecord struct {
	RolName     string
	Attributes  string
	MemberOf    string
	Description string
}

func (a *PostgreSQLAdapter) listRoles() (roles []RoleRecord, err error) {
	ctx := context.Background()
	rows, err := a.conn.QueryContext(ctx, `SELECT r.rolname,
			array_to_string(array_agg(CASE WHEN r.rolsuper THEN 'Superuser' END ||
									CASE WHEN r.rolcreaterole THEN 'Create role' END ||
									CASE WHEN r.rolcreatedb THEN 'Create DB' END ||
									CASE WHEN r.rolcanlogin THEN 'Can login' END ||
									CASE WHEN r.rolreplication THEN 'Replication' END ||
									CASE WHEN r.rolbypassrls THEN 'Bypass RLS' END), ', ') AS attributes,
			array_to_string(ARRAY(SELECT b.rolname
								FROM pg_catalog.pg_auth_members m
								JOIN pg_catalog.pg_roles b ON (m.roleid = b.oid)
								WHERE m.member = r.oid), ', ') as memberof,
			pg_catalog.shobj_description(r.oid, 'pg_authid') AS description
		FROM pg_catalog.pg_roles r
		GROUP BY r.rolname, r.oid
		ORDER BY r.rolname;`)

	defer func() {
		r_err := rows.Close()
		if r_err != nil {
			err = r_err
		}
	}()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var role RoleRecord
		var attributes sql.NullString
		var memberof sql.NullString
		var description sql.NullString
		if err := rows.Scan(&role.RolName, &attributes, &memberof, &description); err != nil {
			return nil, err
		}
		if attributes.Valid {
			role.Attributes = attributes.String
		}
		if memberof.Valid {
			role.MemberOf = memberof.String
		}
		if description.Valid {
			role.Description = description.String
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (a *PostgreSQLAdapter) listDatabases() (databases []DatabaseRecord, err error) {
	// defer db.Close()

	// rows, err := db.QueryContext(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false;")
	ctx := context.Background()
	rows, err := a.conn.QueryContext(ctx, `SELECT
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
	defer func() {
		r_err := rows.Close()
		if r_err != nil {
			err = r_err
		}
	}()

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
			Name:     name,
			Owner:    owner,
			Encoding: encoding,
			Collate:  collate,
			Ctype:    ctype,
		}

		if accessPrivileges.Valid {
			database.AccessPrivileges = accessPrivileges.String
		}
		databases = append(databases, database)
	}

	return databases, nil
}

func (a *PostgreSQLAdapter) listTables() (tables []TableRecord, err error) {
	ctx := context.Background()
	rows, err := a.conn.QueryContext(ctx, `SELECT n.nspname as "Schema",
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
	defer func() {
		r_err := rows.Close()
		if r_err != nil {
			err = r_err
		}
	}()

	for rows.Next() {
		var name string
		var schema string
		var tableType string
		var owner string
		if err := rows.Scan(&schema, &name, &tableType, &owner); err != nil {
			return nil, err
		}

		table := TableRecord{
			Name:   name,
			Schema: schema,
			Type:   tableType,
			Owner:  owner,
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
