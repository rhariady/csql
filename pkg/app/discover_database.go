package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/discovery"
	"github.com/rhariady/csql/pkg/session"
)

type DiscoverDatabase struct{
	instance_list *InstanceList
}

func (d *DiscoverDatabase) GetTitle() string {
	return "Discover Databases"
}

func (d *DiscoverDatabase) GetContent(session *session.Session) tview.Primitive {
	discoveryTypeTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	row := 0
	// for _, disc := range discovery.DiscoveryMap {
	// 	discoveryTypeTable.SetCell(row, 0, tview.NewTableCell(disc.GetLabel()).SetReference(disc).SetExpansion(1))
	// 	row++
	// }
	
	for _, disc := range discovery.GetAllDiscovery() {
		discoveryTypeTable.SetCell(row, 0, tview.NewTableCell(disc.GetLabel()).SetReference(disc).SetExpansion(1))
		row++
	}

	discoveryTypeTable.Select(0, 0)

	discoveryTypeTable.SetDoneFunc(func(key tcell.Key){
		if key == tcell.KeyEscape {
			session.CloseModal()
		}
	})
	discoveryTypeTable.SetSelectedFunc(func(row, col int) {
		session.CloseModal()
		disc := discoveryTypeTable.GetCell(row, 0).Reference.(discovery.IDiscovery)

		view := NewDiscoverDatabaseDetail(d, disc)
		session.ShowModal(view)
	})

	return discoveryTypeTable
}

func (i *DiscoverDatabase) GetKeyBindings() (keybindings []*session.KeyBinding) {
	return
}

func NewDiscoverDatabase(instance_list *InstanceList) *DiscoverDatabase {
	return &DiscoverDatabase{
		instance_list,
	}
}

type DiscoverDatabaseDetail struct{
	*DiscoverDatabase
	discovery discovery.IDiscovery
}

func (d *DiscoverDatabaseDetail) GetTitle() string {
	return fmt.Sprintf("Discover Instances - %s", d.discovery.GetLabel())
}

func (d *DiscoverDatabaseDetail) GetContent(s *session.Session) tview.Primitive {
	form := tview.NewForm()

	d.discovery.GetOptionField(form)

	form.AddButton("Add", func() {
		s.CloseModal()
		s.ShowMessage("Discovering Instance(s)")
		go func() {

			newInstances := d.discovery.DiscoverInstances(form)

			time.Sleep(1 * time.Second)
			
			s.CloseMessageAsync()

			messages := []string{
				"These new instances will be added to config:",
				"",
			}

			for _, newInstance := range newInstances {
				messages = append(messages, newInstance.Name)
			}

			s.ShowAlertAsync(strings.Join(messages, "\n"), func(s *session.Session){
				for _, newInstance := range newInstances {
					s.Config.AddInstance(newInstance)
				}
				s.Config.WriteConfig()
				d.DiscoverDatabase.instance_list.RefreshInstanceTable(s)
			}, func(s *session.Session){
			})
			
			// app.QueueUpdateDraw(func() {
			// 	time.Sleep(1 * time.Second)
			// 	pages.RemovePage("loading-discovery")
			// 	newInstances := tview.NewModal().
			// 		SetText("New instances has been added...").
			// 		AddButtons([]string{"OK"}).
			// 		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// 			pages.RemovePage("new-instances")
			// 		})
			// 	pages.AddPage("new-instances", newInstances, true, true)
			// 	// RefreshInstanceTable(databaseInstanceList)
			// })
		}()
	}).
		AddButton("Cancel", func() {
			s.CloseModal()
		})

	// form.SetBorder(true).SetTitle("Add New Instance(s)")
	// form.SetFieldStyle(tcell.StyleDefault.Background(tcell.ColorGrey).Blink(true).Underline(tcell.ColorWhite))
	// form.SetLabelColor(tcell.ColorDarkGreen)
	// form.SetTitleColor(tcell.ColorDarkGreen)

	return form
}

func (i *DiscoverDatabaseDetail) GetKeyBindings() (keybindings []*session.KeyBinding) {
	return
}

func NewDiscoverDatabaseDetail(parent *DiscoverDatabase, disc discovery.IDiscovery) *DiscoverDatabaseDetail {
	return &DiscoverDatabaseDetail{
		DiscoverDatabase: parent,
		discovery: disc,
	}
}

// func ShowAddDatabasesForm(app *tview.Application, pages *tview.Pages, databaseInstanceList *tview.Table) {

// 	form = tview.NewForm().
// 		AddFormItem(sourceDropDown).
// 		AddButton("Add", func() {
// 			loading := tview.NewModal().
// 				SetText("Discovering instances...")
// 			pages.RemovePage("discover-modal")
// 			pages.AddPage("loading-discovery", loading, true, true)

// 			go func() {

// 				factory.DiscoverInstances(cfg, form)
// 				cfg.WriteConfig()
// 				app.QueueUpdateDraw(func() {
// 					time.Sleep(1 * time.Second)
// 					pages.RemovePage("loading-discovery")
// 					newInstances := tview.NewModal().
// 						SetText("New instances has been added...").
// 						AddButtons([]string{"OK"}).
// 						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
// 							pages.RemovePage("new-instances")
// 						})
// 					pages.AddPage("new-instances", newInstances, true, true)
// 					// RefreshInstanceTable(databaseInstanceList)
// 				})
// 			}()
// 		}).
// 		AddButton("Cancel", func() {
// 			pages.RemovePage("discover-modal")
// 		})

// 	form.SetBorder(true).SetTitle("Add New Instance(s)")
// 	form.SetFieldStyle(tcell.StyleDefault.Background(tcell.ColorGrey).Blink(true).Underline(tcell.ColorWhite))
// 	form.SetLabelColor(tcell.ColorDarkGreen)
// 	form.SetTitleColor(tcell.ColorDarkGreen)

// 	sourceDropDown.SetCurrentOption(0)

// 	modal := centeredModal(form, 0, 0)
// 	pages.AddPage("discover-modal", modal, true, true)
// 	app.SetFocus(form)
// }
