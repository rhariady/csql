package discovery

import (
	"context"
	"strings"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rivo/tview"
)

const (
	GCP     DiscoveryType = "gcp"
)

type GCPDiscovery struct {
}

func NewGCPDiscovery(projectId string) *GCPDiscovery {
	return &GCPDiscovery{
	}
}

func (gcp *GCPDiscovery) DiscoverInstances(form *tview.Form) (newInstances []config.InstanceConfig, err error) {
	projectId := form.GetFormItem(0).(*tview.InputField).GetText()
	instances, err := listGCPInstances(projectId)
	if err != nil {
		return nil, err
	}
	for _, instance := range instances {
		var databaseType string
		var port int
		if strings.HasPrefix(instance.DatabaseInstalledVersion, "POSTGRES") {
			databaseType = "PostgreSQL"
			port = 5432
		} else if strings.HasPrefix(instance.DatabaseInstalledVersion, "MYSQL") {
			databaseType = "MySQL"
			port = 3306
		}

		params := make(map[string]any)
		params["Project ID"] = projectId
		for k, v := range instance.Tags {
			params[k] = v
		}
		newInstance := config.InstanceConfig{
			Name:   instance.Name,
			Source: GCP,
			Host:   instance.IpAddresses[0].IpAddress,
			Port:   port,
			Type:   databaseType,
			Users: []config.UserConfig{},
			Params: params,
		}
		newInstances = append(newInstances, newInstance)
		// cfg.AddInstance(instance.Name, newInstance)
	}
	return
}

func (d *GCPDiscovery) GetLabel() string {
	return "GCP (Auto Discovery)"
}

func (d *GCPDiscovery) GetType() string {
	return GCP
}

func (d *GCPDiscovery) GetInstanceType() string {
	return GCP
}

func (d *GCPDiscovery) GetOptionField(form *tview.Form) {
		form.AddInputField("Project ID", "", 0, nil, nil)
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
	// return []*sqladmin.DatabaseInstance{}, nil
}


