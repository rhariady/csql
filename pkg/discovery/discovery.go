package discovery

import (
	"fmt"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rivo/tview"
)

type DiscoveryType = string

type IDiscovery interface {
	DiscoverInstances(*tview.Form) ([]config.InstanceConfig, error)
	GetType() DiscoveryType
	GetLabel() string
	GetInstanceType() string
	GetOptionField(*tview.Form)
}

var discoveries = []IDiscovery{
	&ManualDiscovery{},
	&GCPDiscovery{},
}

var discoveriesMap map[DiscoveryType]IDiscovery

func GetAllDiscovery() map[DiscoveryType]IDiscovery {
	if discoveriesMap == nil {
		discoveriesMap = make(map[DiscoveryType]IDiscovery)
		for _, discovery := range discoveries {
			discoveriesMap[discovery.GetType()] = discovery
		}
	}
	return discoveriesMap
}

func GetDiscovery(discoveryType DiscoveryType) (IDiscovery, error) {
	discoveriesMap := GetAllDiscovery()
	discovery, found := discoveriesMap[discoveryType]
	if !found {
		return nil, fmt.Errorf("Unknown source: %s", discoveryType)
	}
	return discovery, nil
}
