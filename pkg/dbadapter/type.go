package dbadapter

import (
	"fmt"

	_ "github.com/lib/pq"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/dbadapter/postgresql"
)

type DBType = string

const (
	PostgreSQL  DBType = "PostgreSQL"
)

type IDBAdapter interface{
	Connect(instance *config.InstanceConfig, userConfig *config.UserConfig)
	// ListDatabases(instance *config.InstanceConfig, userConfig *config.UserConfig) ([]DatabaseRecord, error)
	// RunShell(instance *config.InstanceConfig, user *config.UserConfig, username string)
}

func GetDBAdapter(dbType DBType) (IDBAdapter, error) {
	if dbType == PostgreSQL {
		return &postgresql.PostgreSQLDBAdapter{}, nil
	}
	return nil, fmt.Errorf("Unknown DB Type")
}

