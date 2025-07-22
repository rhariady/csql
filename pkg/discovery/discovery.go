package discovery

import (
	"context"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rivo/tview"
)

type DiscoveryType = string

const (
	Manual  DiscoveryType = "Manual"
	GCP     DiscoveryType = "GCP (Auto Discovery)"
)

var DiscoveryList = map[DiscoveryType]IDiscovery{
	Manual: &ManualDiscovery{},
	GCP: &GCPDiscovery{},
}

type IDiscovery interface{
	DiscoverInstances(*config.Config, *tview.Form)
	GetLabel() string
	GetInstanceType() string
	GetOptionField(*tview.Form)
}

type ManualDiscovery struct {
}

func NewManualDiscovery(name, host string, port int) *ManualDiscovery {
	return &ManualDiscovery{
	}
}

func (d *ManualDiscovery) DiscoverInstances(cfg *config.Config, form *tview.Form) {
	instanceName := form.GetFormItem(1).(*tview.InputField).GetText()
	host := form.GetFormItem(2).(*tview.InputField).GetText()
	portStr := form.GetFormItem(3).(*tview.InputField).GetText()
	
	newInstance := config.InstanceConfig{
		Name:   instanceName,
		Host:   host,
		Users: map[string]config.UserConfig{},
		Params: map[string]interface{}{
			"discovery": string(Manual),
			"port": portStr,
		},
	}
	cfg.AddInstance(instanceName, newInstance)
}

func (d *ManualDiscovery) GetLabel() string {
	return "Manual"
}

func (d *ManualDiscovery) GetInstanceType() string {
	return "manual"
}

func (d *ManualDiscovery) GetOptionField(form *tview.Form) {
		form.AddInputField("Name", "", 0, nil, nil)
		form.AddInputField("Host", "", 0, nil, nil)
		form.AddInputField("Port", "", 0, nil, nil)
}


type GCPDiscovery struct {
}

func NewGCPDiscovery(projectId string) *GCPDiscovery {
	return &GCPDiscovery{
	}
}

func (gcp *GCPDiscovery) DiscoverInstances(cfg *config.Config, form *tview.Form) {
	projectId := form.GetFormItem(1).(*tview.InputField).GetText()
	instances, err := listGCPInstances(projectId)
	if err != nil {
		panic(err)
	}

	for _, instance := range instances {
		newInstance := config.InstanceConfig{
			Name:   instance.Name,
			Host:   instance.IpAddresses[0].IpAddress,
			Users: map[string]config.UserConfig{},
			Params: map[string]interface{}{
				"discovery": string(GCP),
				"project_id": projectId,
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
}

