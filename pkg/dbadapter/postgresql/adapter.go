package postgresql

import (
	"fmt"
	"os"
	"os/exec"

	_ "github.com/lib/pq"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/auth"
	"github.com/rhariady/csql/pkg/session"
)

type PostgreSQLDBAdapter struct {
	// session *session.Session
	instance *config.InstanceConfig
	user *config.UserConfig

}

func (a *PostgreSQLDBAdapter) Connect(session *session.Session, instance *config.InstanceConfig, user *config.UserConfig) {
	// a.session = session
	a.instance = instance
	a.user = user
	
	// databaseList := NewDatabaseList(instance, user)
	databaseList := NewDatabaseList(a)
	session.SetView(databaseList)
}

// func (a *PostgreSQLDBAdapter) Connect {
// }

func (a *PostgreSQLDBAdapter) RunShell(instance *config.InstanceConfig, user *config.UserConfig, dbname string) {
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
