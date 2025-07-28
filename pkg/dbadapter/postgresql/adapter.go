package postgresql

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	_ "github.com/lib/pq"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/auth"
	"github.com/rhariady/csql/pkg/session"
)

type PostgreSQLAdapter struct {
	// session *session.Session
	instance *config.InstanceConfig
	user *config.UserConfig
	database string
	conn *sql.DB
}

func (a *PostgreSQLAdapter) Connect(session *session.Session, instance *config.InstanceConfig, user *config.UserConfig) error {
	// a.session = session
	a.instance = instance
	a.user = user
	a.database = user.DefaultDatabase

	host := instance.Host
	port := instance.Port

	authConfig, err := auth.GetAuth(user.AuthType, user.AuthParams)
	if err != nil {
		return err
	}

	dbName := user.DefaultDatabase
	if dbName == "" {
		dbName = "postgres"
	}

	password := authConfig.GetCredential()
	connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", user.Username, password, host, port, dbName)

	a.conn, err = sql.Open("postgres", connectionUri)
	if err != nil {
		return err
	}

	// databaseList := NewDatabaseList(instance, user)
	databaseList := NewDatabaseList(a)
	session.SetView(databaseList)

	return nil
}

func (a *PostgreSQLAdapter) Close() error {
	if a.conn != nil {
		a.conn.Close()
	}

	return nil
}

func (a *PostgreSQLAdapter) RunShell(instance *config.InstanceConfig, user *config.UserConfig, dbname string) {
			authConfig, err := auth.GetAuth(user.AuthType, user.AuthParams)
			if err != nil {
				panic(err)
			}

			password := authConfig.GetCredential()
			connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", user.Username, password, instance.Host, instance.Port, dbname)

			fmt.Println("Connecting to:", connectionUri)
			cmd := exec.Command("psql", connectionUri)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("Error:", err)
			}
}
