package discovery

import (
	"github.com/rhariady/csql/pkg/config"
	"github.com/rivo/tview"
)

type DiscoveryType = string

type IDiscovery interface{
	DiscoverInstances(*tview.Form) []config.InstanceConfig
	GetLabel() string
	GetInstanceType() string
	GetOptionField(*tview.Form)
}

func GetAllDiscovery() []IDiscovery {
	discoveries := []IDiscovery{
		&ManualDiscovery{},
		&GCPDiscovery{},
	}
	return discoveries
}

