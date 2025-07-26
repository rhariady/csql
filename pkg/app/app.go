package app

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/auth"
	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/dbadapter"
	"github.com/rhariady/csql/pkg/discovery"
)

var cfg *config.Config

func ShowUserSelection(app *tview.Application, pages *tview.Pages, mainTable *tview.Table, instanceName string) {
	userTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	userRow := 0
	for _, user := range cfg.Instances[instanceName].Users {
		userTable.SetCell(userRow, 0, tview.NewTableCell(user.Username))
		userTable.SetCell(userRow, 1, tview.NewTableCell(fmt.Sprintf("[auth=%s]", user.AuthType)).SetExpansion(1))
		userRow++
	}

	userTable.Select(0, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			pages.RemovePage("userSelectionModal")
			// app.SetRoot(mainTable, true) // Go back to instance table
		}
	}).SetSelectedFunc(func(row int, column int) {
		userName := userTable.GetCell(row, 0).Text
		ShowDatabaseList(app, pages, instanceName, userName, userTable)
	})

	modalFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(userTable, 0, 1, true).
		AddItem(tview.NewTextView().SetText("Press Esc to go back, 'a' to add user").SetTextAlign(tview.AlignCenter), 1, 1, false)
	modalFlex.SetBorder(true).SetTitle("Select user")

	// Center the modal
	centeredModal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modalFlex, 0, 1, true).
			AddItem(nil, 0, 1, false), 0, 1, false).
		AddItem(nil, 0, 1, false)

	pages.AddPage("userSelectionModal", centeredModal, true, true)
	// pages.SwitchToPage("userSelectionModal")
	app.SetFocus(userTable)

	centeredModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'a' {
			ShowAddUserForm(app, pages, mainTable, instanceName)
			return nil // Consume the event
		}
		return event
	})
}

func ShowDatabaseList(app *tview.Application, pages *tview.Pages, instanceName string, userName string, userTable *tview.Table) {
	// Table for databases
	databaseTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	databaseTable.SetCell(0, 0, tview.NewTableCell("Name").SetSelectable(false))
	databaseTable.SetCell(0, 1, tview.NewTableCell("Owner").SetSelectable(false))
	databaseTable.SetCell(0, 2, tview.NewTableCell("Encoding").SetSelectable(false))
	databaseTable.SetCell(0, 3, tview.NewTableCell("Collate").SetSelectable(false))
	databaseTable.SetCell(0, 4, tview.NewTableCell("Ctype").SetSelectable(false))
	databaseTable.SetCell(0, 5, tview.NewTableCell("Access Privileges").SetSelectable(false))
	databaseTable.SetCell(1, 0, tview.NewTableCell("Loading databases..."))

	// Get databases
	go func() {
		instance := cfg.Instances[instanceName]
		user, _ := instance.GetUserConfig(userName)
		dbAdapter, _ := dbadapter.GetDBAdapter(instance.Type)
		databases, err := dbAdapter.ListDatabases(&instance, user)
		if err != nil {
			// Show an error modal
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error loading databases: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					pages.RemovePage("errorModal")
					app.SetFocus(userTable)
				})
			pages.AddPage("errorModal", errorModal, true, true)
			app.SetFocus(errorModal)
		}

		// Populate table
		for i, db := range databases {
			databaseTable.SetCell(i+1, 0, tview.NewTableCell(db.Name))
			databaseTable.SetCell(i+1, 1, tview.NewTableCell(db.Owner))
			databaseTable.SetCell(i+1, 2, tview.NewTableCell(db.Encoding))
			databaseTable.SetCell(i+1, 3, tview.NewTableCell(db.Collate))
			databaseTable.SetCell(i+1, 4, tview.NewTableCell(db.Ctype))
			databaseTable.SetCell(i+1, 5, tview.NewTableCell(db.AccessPrivileges))
		}

		// On selection, connect to DB
		databaseTable.SetSelectedFunc(func(row int, column int) {
			if row == 0 { // Skip header
				return
			}
			dbName := databaseTable.GetCell(row, 0).Text
			app.Stop() // Stop the tview app to hand over to psql

			instance := cfg.Instances[instanceName]
			user, _ := instance.GetUserConfig(userName)
			dbAdapter, _ := dbadapter.GetDBAdapter(instance.Type)
			dbAdapter.RunShell(&instance, user, dbName)

		})

		app.Draw()
	}()

	flex := tview.NewFlex().AddItem(databaseTable, 0, 1, true)
	flex.SetBorder(true).SetTitle("Databases")

	// Go back on escape
	databaseTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			pages.SwitchToPage("mainTable")
		}
	})

	pages.AddPage("databaseTable", flex, true, true)
	app.SetFocus(databaseTable)
}

// Helper function to center a primitive
func centeredModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func ShowAddUserForm(app *tview.Application, pages *tview.Pages, mainTable *tview.Table, instanceName string) {
	var form *tview.Form
	auth_type := tview.NewDropDown().
		SetLabel("Auth Type").
		SetListStyles(tcell.StyleDefault.Background(tcell.ColorNone), tcell.StyleDefault.Background(tcell.ColorGrey)).
		SetFocusedStyle(tcell.StyleDefault.Background(tcell.ColorGrey)).
		SetPrefixStyle(tcell.StyleDefault.Background(tcell.ColorGrey))

	//SetFieldStyle(tcell.StyleDefault.Background(tcell.ColorGrey))

	for authType, authConfig := range auth.AuthList {
		auth_type.AddOption(authType, func() {
			for form.GetFormItemCount() > 2 {
				form.RemoveFormItem(2)
			}
			authConfig.GetFormInput(form)
		})
	}

	form = tview.NewForm().
		AddInputField("Username", "", 0, nil, nil).
		AddFormItem(auth_type).
		AddButton("Add User", func() {
			username := form.GetFormItem(0).(*tview.InputField).GetText()
			_, authType := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()

			authAdapter, _ := auth.GetAuth(authType, nil)
			authParams := authAdapter.ParseFormInput(form)
			newUser := config.UserConfig{
				Username:   username,
				AuthType:   authType,
				AuthParams: authParams,
			}

			cfg.AddInstanceUser(instanceName, newUser)
			cfg.WriteConfig()

			pages.RemovePage("addUserModal")
			pages.RemovePage("userSelectionModal")
			ShowUserSelection(app, pages, mainTable, instanceName)
		}).
		AddButton("Cancel", func() {
			pages.RemovePage("addUserModal")
		})

	form.SetBorder(true).SetTitle("Add New User")
	// form.SetFieldBackgroundColor(tcell.ColorDarkGreen)
	fieldStyle := tcell.StyleDefault.
		Background(tcell.ColorGrey).
		Blink(true).
		Underline(tcell.ColorWhite)
	form.SetFieldStyle(fieldStyle)
	form.SetLabelColor(tcell.ColorDarkGreen)
	form.SetTitleColor(tcell.ColorDarkGreen)

	auth_type.SetCurrentOption(0)

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 1, true).
			AddItem(nil, 0, 1, false), 0, 1, false).
		AddItem(nil, 0, 1, false)

	pages.AddPage("addUserModal", modal, true, true)
	app.SetFocus(form)
}

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
					RefreshInstanceTable(databaseInstanceList)
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

func RefreshInstanceTable(table *tview.Table) {
	table.Clear()
	table.SetCell(0, 0, tview.NewTableCell("Name").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Type").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Host").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 3, tview.NewTableCell("Port").SetExpansion(1).SetSelectable(false)).
		SetCell(0, 4, tview.NewTableCell("Params").SetExpansion(1).SetSelectable(false))

	table.SetWrapSelection(true, true)
	row := 1
	for name, instance := range cfg.Instances {
		table.SetCell(row, 0, tview.NewTableCell(name))
		table.SetCell(row, 1, tview.NewTableCell(instance.Type))
		table.SetCell(row, 2, tview.NewTableCell(instance.Host))
		table.SetCell(row, 3, tview.NewTableCell(fmt.Sprint(instance.Port)))
		table.SetCell(row, 4, tview.NewTableCell(fmt.Sprint(instance.Params)))
		row++
	}
}

func NewApplication(config *config.Config) *tview.Application {
	cfg = config
	app := tview.NewApplication()

	pages := tview.NewPages()

	// Create the main database table
	databaseInstanceList := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	RefreshInstanceTable(databaseInstanceList)

	// Set the selected function for the table (triggered by Enter key)
	databaseInstanceList.SetSelectedFunc(func(row int, column int) {
		if row == 0 { // Skip header row
			return
		}
		instanceName := databaseInstanceList.GetCell(row, 0).Text
		ShowUserSelection(app, pages, databaseInstanceList, instanceName)
	})

	// Set input capture for 'a' key to trigger the same selection logic
	databaseInstanceList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'a' {
			ShowAddDatabasesForm(app, pages, databaseInstanceList)
			return nil // Consume the event
		}
		return event
	})

	flex := tview.NewFlex().AddItem(databaseInstanceList, 0, 1, true)
	flex.SetBorder(true).SetTitle("Instances")

	logo := tview.NewTextView().SetText(`
   ___________ ____    __
  / ____/ ___// __ \  / /
 / /    \__ \/ / / / / /
/ /___ ___/ / /_/ / / /___
\____//____/\___\_\/_____/`)

	keyBindings := []struct {
		Key     string
		Purpose string
	}{
		{"q", "quit"},
		{"a", "add instance"},
		{"enter", "select"},
	}

	keyLegend := tview.NewGrid().
		SetRows(1, 1, 1, 1, 1, 1)

	for i, binding := range keyBindings {
		x := i / 6
		y := i % 6
		keyLegend.AddItem(tview.NewTextView().SetText(fmt.Sprintf("<%s> %s", binding.Key, binding.Purpose)), y, x, 1, 1, 0, 0, false)
	}

	header := tview.NewFlex().
		AddItem(keyLegend, 0, 1, false).
		AddItem(tview.NewBox(), 0, 2, true).
		AddItem(logo, 28, 0, false)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 6, 0, false).
		AddItem(flex, 0, 1, true)

	pages.AddPage("main", mainFlex, true, true)

	app.SetRoot(pages, true)

	return app
}
