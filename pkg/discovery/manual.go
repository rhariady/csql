package discovery

import (
	"strconv"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rivo/tview"
)

const (
	Manual  DiscoveryType = "manual"
)

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
	form.AddDropDown("Database Type", []string{"PostgreSQL", "MySQL"}, 0, nil)
	form.AddInputField("Name", "", 0, nil, nil)
	form.AddInputField("Host", "", 0, nil, nil)
	form.AddInputField("Port", "", 0, nil, nil)
}
