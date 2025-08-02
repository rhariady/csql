package discovery

import (
	"context"
	"strconv"
	"strings"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rivo/tview"
)

type DiscoveryType = string

const (
	Manual  DiscoveryType = "manual"
	GCP     DiscoveryType = "gcp"
)

// var DiscoveryMap = map[DiscoveryType]IDiscovery{
// 	Manual: &ManualDiscovery{},
// 	GCP: &GCPDiscovery{},
// }

func GetAllDiscovery() []IDiscovery {
	// discoveries := make([]IDiscovery, 0, len(DiscoveryMap))
	// for _, d := range DiscoveryMap {
	// 	discoveries = append(discoveries, d)
	// }
	discoveries := []IDiscovery{
		&ManualDiscovery{},
		&GCPDiscovery{},
	}
	return discoveries
}

type IDiscovery interface{
	DiscoverInstances(*tview.Form) []config.InstanceConfig
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

func (d *ManualDiscovery) DiscoverInstances(form *tview.Form) (newInstances []config.InstanceConfig) {
	_, databaseType := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
	instanceName := form.GetFormItem(1).(*tview.InputField).GetText()
	host := form.GetFormItem(2).(*tview.InputField).GetText()
	port, _ := strconv.Atoi(form.GetFormItem(3).(*tview.InputField).GetText())
	
	newInstance := config.InstanceConfig{
		Name:   instanceName,
		Host:   host,
		Port:   port,
		Type:   databaseType,
		Users: []config.UserConfig{},
		Params: map[string]interface{}{
			"discovery": string(Manual),
		},
	}
	// cfg.AddInstance(instanceName, newInstance)
	newInstances = append(newInstances, newInstance)
	
	return
}

func (d *ManualDiscovery) GetLabel() string {
	return "Manual"
}

func (d *ManualDiscovery) GetInstanceType() string {
	return Manual
}

func (d *ManualDiscovery) GetOptionField(form *tview.Form) {
	form.AddDropDown("Database Type", []string{"PostgreSQL"}, 0, nil)
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

func (gcp *GCPDiscovery) DiscoverInstances(form *tview.Form) (newInstances []config.InstanceConfig) {
	projectId := form.GetFormItem(1).(*tview.InputField).GetText()
	instances, err := listGCPInstances(projectId)
	if err != nil {
		panic(err)
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
		newInstance := config.InstanceConfig{
			Name:   instance.Name,
			Host:   instance.IpAddresses[0].IpAddress,
			Port:   port,
			Type:   databaseType,
			Users: []config.UserConfig{},
			Params: map[string]interface{}{
				"discovery": string(GCP),
				"project_id": projectId,
			},
		}
		newInstances = append(newInstances, newInstance)
		// cfg.AddInstance(instance.Name, newInstance)
	}
	return
}

func (d *GCPDiscovery) GetLabel() string {
	return "GCP (Auto Discovery)"
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
}

