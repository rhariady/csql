package discovery

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/config"
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
		Source: Manual,
		Host:   host,
		Port:   port,
		Type:   databaseType,
		Users: []config.UserConfig{},
		Params: map[string]interface{}{},
	}
	// cfg.AddInstance(instanceName, newInstance)
	newInstances = append(newInstances, newInstance)
	
	return
}

func (d *ManualDiscovery) GetLabel() string {
	return "Manual"
}

func (d *ManualDiscovery) GetType() string {
	return Manual
}

func (d *ManualDiscovery) GetInstanceType() string {
	return Manual
}

func (d *ManualDiscovery) GetOptionField(form *tview.Form) {
	database_type := tview.NewDropDown().SetLabel("Database Type").
		AddOption("PostgreSQL", func() {
			port_field := form.GetFormItemByLabel("Port").(*tview.InputField)
			port_field.SetText("5432")
		}).
		AddOption("MySQL", func() {
			port_field := form.GetFormItemByLabel("Port").(*tview.InputField)
			port_field.SetText("3306")
		})

	database_type.SetListStyles(tcell.StyleDefault.Background(tcell.ColorGray), tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorGreen)).
		SetFocusedStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorGreen)).
		SetPrefixStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorGreen))

	// form.AddDropDown("Database Type", []string{"PostgreSQL", "MySQL"}, 0, nil)
	form.AddFormItem(database_type)
	form.AddInputField("Name", "", 0, nil, nil)
	form.AddInputField("Host", "", 0, nil, nil)
	form.AddInputField("Port", "", 0, nil, nil)

	database_type.SetCurrentOption(0)
}
