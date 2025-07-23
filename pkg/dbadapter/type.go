package dbadapter

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	_ "github.com/lib/pq"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/auth"
)

type DBType = string

const (
	PostgreSQL  DBType = "PostgreSQL"
)

type IDBAdapter interface{
	ListDatabases(instance *config.InstanceConfig, username string) ([]DatabaseRecord, error)
	RunShell(instance *config.InstanceConfig, dbname string, username string)
}

func GetDBAdapter(dbType DBType) (IDBAdapter, error) {
	if dbType == PostgreSQL {
		return &PostgreSQLDBAdapter{}, nil
	}
	return nil, fmt.Errorf("Unknown DB Type")
}

type PostgreSQLDBAdapter struct {
}

type DatabaseRecord struct {
	Name string
	Owner string
	Encoding string
	Collate string
	Ctype string
	AccessPrivileges string
}


func (a *PostgreSQLDBAdapter) ListDatabases(instance *config.InstanceConfig, username string) ([]DatabaseRecord, error) {
	user := username
	host := instance.Host
	port := instance.Port
	dbname := "postgres" // Connect to a default database to list others

	authConfig, err := auth.GetAuth(instance.Users[username].DefaultAuth, instance.Users[username].Auth)
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
	
func (a *PostgreSQLDBAdapter) RunShell(instance *config.InstanceConfig, dbname string, username string) {
			authConfig, err := auth.GetAuth(instance.Users[username].DefaultAuth, instance.Users[username].Auth)
			if err != nil {
				panic(err)
			}

			password := authConfig.GetCredential()
			connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", username, password, instance.Host, instance.Port, dbname)

			fmt.Println("Connecting to:", connectionUri)
			cmd := exec.Command("psql", connectionUri)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("Error:", err)
			}
}
