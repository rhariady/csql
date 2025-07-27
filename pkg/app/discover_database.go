package app

import (	
	"time"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/discovery"
)

func ShowAddDatabasesForm(app *tview.Application, pages *tview.Pages, databaseInstanceList *tview.Table) {
	var form *tview.Form
	sourceDropDown := tview.NewDropDown().
		SetLabel("Source").
		SetListStyles(tcell.StyleDefault.Background(tcell.ColorNone), tcell.StyleDefault.Background(tcell.ColorGrey)).
		SetFocusedStyle(tcell.StyleDefault.Background(tcell.ColorGrey)).
		SetPrefixStyle(tcell.StyleDefault.Background(tcell.ColorGrey))

	allDiscovery := discovery.GetAllDiscovery()
	for _, discovery := range allDiscovery {
		sourceDropDown.AddOption(discovery.GetLabel(), func() {
			for form.GetFormItemCount() > 1 {
				form.RemoveFormItem(1)
			}
			discovery.GetOptionField(form)
		})
	}

	form = tview.NewForm().
		AddFormItem(sourceDropDown).
		AddButton("Add", func() {
			_, selectedSource := sourceDropDown.GetCurrentOption()

			var factory discovery.IDiscovery
			for _, d := range allDiscovery {
				if d.GetLabel() == selectedSource {
					factory = d
					break
				}
			}

			loading := tview.NewModal().
				SetText("Discovering instances...")
			pages.RemovePage("discover-modal")
			pages.AddPage("loading-discovery", loading, true, true)

			go func() {

				factory.DiscoverInstances(cfg, form)
				cfg.WriteConfig()
				app.QueueUpdateDraw(func() {
					time.Sleep(1 * time.Second)
					pages.RemovePage("loading-discovery")
					newInstances := tview.NewModal().
						SetText("New instances has been added...").
						AddButtons([]string{"OK"}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							pages.RemovePage("new-instances")
						})
					pages.AddPage("new-instances", newInstances, true, true)
					// RefreshInstanceTable(databaseInstanceList)
				})
			}()
		}).
		AddButton("Cancel", func() {
			pages.RemovePage("discover-modal")
		})

	form.SetBorder(true).SetTitle("Add New Instance(s)")
	form.SetFieldStyle(tcell.StyleDefault.Background(tcell.ColorGrey).Blink(true).Underline(tcell.ColorWhite))
	form.SetLabelColor(tcell.ColorDarkGreen)
	form.SetTitleColor(tcell.ColorDarkGreen)

	sourceDropDown.SetCurrentOption(0)

	modal := centeredModal(form, 0, 0)
	pages.AddPage("discover-modal", modal, true, true)
	app.SetFocus(form)
}
