package discovery

import (
	"context"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/rhariady/csql/pkg/config"	
)

const (
	Manual config.SourceType = "Manual"
	GCP    config.SourceType = "GCP"
)

type Discovery interface{
	DiscoverInstances(*config.Config)
	GetLabel() string
	GetInstanceType() string
}

type ManualDiscovery struct {
	name string
	host string
	port int
}

func NewManualDiscovery(name, host string, port int) *ManualDiscovery {
	return &ManualDiscovery{
		name: name,
		host: host,
		port: port,
	}
}

func (d *ManualDiscovery) DiscoverInstances(cfg *config.Config) {
	newInstance := config.InstanceConfig{
			Name:   d.name,
			Host:   d.host,
			Source: Manual,
			Users: map[string]config.UserConfig{},
			Params: map[string]interface{}{
				"port": d.port,
			},
	}
	cfg.AddInstance(d.name, newInstance)
}

func (d *ManualDiscovery) GetLabel() string {
	return "Manual"
}

func (d *ManualDiscovery) GetInstanceType() string {
	return "manual"
}

type GCPDiscovery struct {
	projectId string
}

func NewGCPDiscovery(projectId string) *GCPDiscovery {
	return &GCPDiscovery{
		projectId: projectId,
	}
}

func (gcp *GCPDiscovery) DiscoverInstances(cfg *config.Config) {
	instances, err := listGCPInstances(gcp.projectId)
	if err != nil {
		panic(err)
	}

	for _, instance := range instances {
		newInstance := config.InstanceConfig{
			Name:   instance.Name,
			Host:   instance.IpAddresses[0].IpAddress,
			Source: GCP,
			Users: map[string]config.UserConfig{},
			Params: map[string]interface{}{
				"project_id": gcp.projectId,
			},
		}
		cfg.AddInstance(instance.Name, newInstance)
	}
}

func (d *GCPDiscovery) GetLabel() string {
	return "GCP (Auto Discovery)"
}

func (d *GCPDiscovery) GetInstanceType() string {
	return "gcp"
}

func listGCPInstances(projectId string) ([]*sqladmin.DatabaseInstance, error) {
	ctx := context.Background()
	service, err := sqladmin.NewService(ctx)
	if err != nil {
		return nil, err
	}

	instances, err := service.Instances.List(projectId).Do()
	if err != nil {
		return nil, err
	}
	return instances.Items, nil
}

