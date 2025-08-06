package dbadapter

import (
	"fmt"

	_ "github.com/lib/pq"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/session"

	"github.com/rhariady/csql/pkg/dbadapter/postgresql"
)

type DBType = string

const (
	PostgreSQL  DBType = "PostgreSQL"
)

type IDBAdapter interface{
	Connect(session *session.Session,  instance *config.InstanceConfig, userConfig *config.UserConfig, database string) error
	Close() error
}

func GetDBAdapter(dbType DBType) (IDBAdapter, error) {
	var adapter IDBAdapter
	if dbType == PostgreSQL {
		adapter = &postgresql.PostgreSQLAdapter{}
		RegisterDBAdapter(adapter)
		return adapter, nil
	}
	return nil, fmt.Errorf("Unsupported DB Type")
}

var adapters []IDBAdapter

func RegisterDBAdapter(adapter IDBAdapter) {
		if adapters == nil {
			adapters = make([]IDBAdapter, 0)
		}

		adapters = append(adapters, adapter)	
}

func CloseAllAdapter() {
	for _, adapter := range adapters {
		if adapter != nil {
			adapter.Close()
		}
	}
}

