package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	vault "github.com/hashicorp/vault/api"
	_ "github.com/lib/pq"
	"github.com/mattn/go-isatty"
	"github.com/rivo/tview"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

var config *Config

func main() {
	var app *tview.Application
	var pages *tview.Pages

	CheckConfigFile()
	var err error
	config, err = GetConfig()
	if err != nil {
		panic(err)
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		panic("This application is intended to be run in an interactive terminal.")
	} else {
		app = tview.NewApplication()

		// Create the main database table
		databaseInstanceList := tview.NewTable().
			SetBorders(false).
			SetSelectable(true, false)

		// Populate the table with database instances
		databaseInstanceList.SetCell(0, 0, tview.NewTableCell("Name").SetSelectable(false)).
			SetCell(0, 1, tview.NewTableCell("Project ID").SetSelectable(false)).
			SetCell(0, 2, tview.NewTableCell("Host").SetSelectable(false))

		databaseInstanceList.SetWrapSelection(true, true)

		row := 1
		for name, instance := range config.Instances {
			databaseInstanceList.SetCell(row, 0, tview.NewTableCell(name))
			databaseInstanceList.SetCell(row, 1, tview.NewTableCell(instance.ProjectID))
			databaseInstanceList.SetCell(row, 2, tview.NewTableCell(instance.Host))
			row++
		}

		// Set the selected function for the table (triggered by Enter key)
		databaseInstanceList.SetSelectedFunc(func(row int, column int) {
			if row == 0 { // Skip header row
				return
			}
			instanceName := databaseInstanceList.GetCell(row, 0).Text
			showUserSelection(app, pages, databaseInstanceList, instanceName)
		})

		// Set input capture for 'a' key to trigger the same selection logic
		databaseInstanceList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Rune() == 'a' {
				discoverDatabases()
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
		keyLegend := tview.NewTextView().SetText(`(q) quit
(a) add instance
(enter) select`)

		header := tview.NewFlex().
			AddItem(keyLegend, 0, 1, false).
			AddItem(tview.NewBox(), 0, 2, true).
			AddItem(logo, 28, 0, false)

		mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(header, 6, 0, false).
			AddItem(flex, 0, 1, true)


		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			}
			return event
		})

		pages = tview.NewPages().
			AddPage("main", mainFlex, true, true)
			//AddPage("userModal", centeredModal(modalFlex, 60, 20), true, true)

		//app.SetRoot(pages, true).SetFocus(userTable)
		//pages.SwitchToPage("userModal")

		if err := app.SetRoot(pages, true).Run(); err != nil {
			panic(err)
		}
	}
}

func showUserSelection(app *tview.Application, pages *tview.Pages, mainTable *tview.Table, instanceName string) {
	userTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	userRow := 0
	for name, user := range config.Instances[instanceName].Users {
		userTable.SetCell(userRow, 0, tview.NewTableCell(name))
		userTable.SetCell(userRow, 1, tview.NewTableCell(fmt.Sprintf("[auth=%s]", user.DefaultAuth)).SetExpansion(1))
		userRow++
	}

	userTable.Select(0, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			pages.RemovePage("userSelectionModal")
			// app.SetRoot(mainTable, true) // Go back to instance table
		}
	}).SetSelectedFunc(func(row int, column int) {
		userName := userTable.GetCell(row, 0).Text
		showDatabaseList(app, pages, instanceName, userName, userTable)
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
			showAddUserForm(app, pages, mainTable, instanceName)
			return nil // Consume the event
		}
		return event
	})
}

func showDatabaseList(app *tview.Application, pages *tview.Pages, instanceName string, userName string, userTable *tview.Table) {
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
	projectID := config.Instances[instanceName].ProjectID
	ctx := context.Background()


	go func() {
		databases, err := ListDatabases(ctx, projectID, instanceName, userName)
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

			user := config.Instances[instanceName].Users[userName].Username
			host := config.Instances[instanceName].Host
			port := 5432

			authConfig, err := NewAuthConfig(config.Instances[instanceName].Users[userName].DefaultAuth, config.Instances[instanceName].Users[userName].Auth)
			if err != nil {
				panic(err)
			}

			password := authConfig.GetCredential()
			connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", user, password, host, port, dbName)

			fmt.Println("Connecting to:", connectionUri)
			cmd := exec.Command("psql", connectionUri)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("Error:", err)
			}
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

func showAddUserForm(app *tview.Application, pages *tview.Pages, mainTable *tview.Table, instanceName string) {
	var form *tview.Form
	auth_type := tview.NewDropDown().
		SetLabel("Auth Type").
		SetListStyles(tcell.StyleDefault.Background(tcell.ColorNone), tcell.StyleDefault.Background(tcell.ColorGrey)).
		SetFocusedStyle(tcell.StyleDefault.Background(tcell.ColorGrey)).
		SetPrefixStyle(tcell.StyleDefault.Background(tcell.ColorGrey))

		//SetFieldStyle(tcell.StyleDefault.Background(tcell.ColorGrey))


	auth_type.AddOption("local", func() {
		for form.GetFormItemCount() > 2 {
			form.RemoveFormItem(2)
		}

		form.AddInputField("Password", "", 0, nil, nil)
	})

	auth_type.AddOption("vault", func() {
		item_count := form.GetFormItemCount()
		for i := 2; i < item_count; i++ {
			form.RemoveFormItem(i)
		}

		form.
			AddInputField("Vault Address", "", 0, nil, nil).
			AddInputField("Vault Mount Path", "", 0, nil, nil).
			AddInputField("Vault Secret Path", "", 0, nil, nil).
			AddInputField("Vault Secret Key", "", 0, nil, nil)
	})

	form = tview.NewForm().
		AddInputField("Username", "", 0, nil, nil).
		AddFormItem(auth_type).
		AddButton("Add User", func() {
			username := form.GetFormItem(0).(*tview.InputField).GetText()
			_, authType := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()

			var newUser UserConfig
			if authType == "vault" {
				vaultAddress := form.GetFormItem(2).(*tview.InputField).GetText()
				vaultMountPath := form.GetFormItem(3).(*tview.InputField).GetText()
				vaultSecretPath := form.GetFormItem(4).(*tview.InputField).GetText()
				vaultSecretKey := form.GetFormItem(5).(*tview.InputField).GetText()
				newUser = UserConfig{
					Username:    username,
					DefaultAuth: "vault",
					Auth: map[string]interface{}{
						"vault": map[string]string{
							"address":     vaultAddress,
							"mount_path":  vaultMountPath,
							"secret_path": vaultSecretPath,
							"secret_key":  vaultSecretKey,
						},
					},
				}
			} else if authType == "local" {
				password := form.GetFormItem(2).(*tview.InputField).GetText()
				newUser = UserConfig{
					Username:    username,
					DefaultAuth: "local",
					Auth: map[string]interface{}{
						"local": map[string]string{
							"password": password,
						},
					},
				}
			}

			config.Instances[instanceName].Users[username] = newUser
			config.WriteConfig()

			pages.RemovePage("addUserModal")
			pages.RemovePage("userSelectionModal")
			showUserSelection(app, pages, mainTable, instanceName)
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

func discoverDatabases() {
	app := tview.NewApplication()
	var form *tview.Form
	form = tview.NewForm().
		AddDropDown("Select Discovery Source", []string{"GCP", "AWS", "Azure", "Manual (VM)"}, 0, nil).
		AddInputField("GCP Project ID", "", 20, nil, nil).
		AddButton("Discover", func() {
			_, source := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
			projectId := form.GetFormItem(1).(*tview.InputField).GetText()

			if source == "GCP" {
				modal := tview.NewModal().
					SetText("Discovering instances...").
					AddButtons([]string{"Cancel"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						if buttonLabel == "Cancel" {
							app.Stop()
						}
					})

				go func() {
					ctx := context.Background()
					DiscoverInstances(config, ctx, projectId)
					app.QueueUpdateDraw(func() {
						app.Stop()
						fmt.Println("Discovery complete. Please restart the application to see the new instances.")
					})
				}()

				if err := app.SetRoot(modal, false).Run(); err != nil {
					panic(err)
				}
			} else {
				app.Stop()
				fmt.Println(source, "discovery not yet implemented.")
			}
		}).
		AddButton("Quit", func() {
			app.Stop()
		})
	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}

func DiscoverInstances(config *Config, ctx context.Context, projectId string) {
	fmt.Printf("Project %v\n", projectId)
	instances, err := ListInstances(ctx, projectId)
	if err != nil {
		panic(err)
	}

	for _, instance := range instances {
		fmt.Printf("Found an Instance %v: \n", instance.Name, instance)
		newInstance := InstanceConfig{
			Name:      instance.Name,
			ProjectID: projectId,
			Host:      instance.IpAddresses[0].IpAddress,
		}
		config.AddInstance(instance.Name, newInstance)
	}
}

func ListInstances(ctx context.Context, projectId string) ([]*sqladmin.DatabaseInstance, error) {
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

type DatabaseRecord struct {
	Name string
	Owner string
	Encoding string
	Collate string
	Ctype string
	AccessPrivileges string
}

func ListDatabases(ctx context.Context, projectID string, instanceName string, userName string) ([]DatabaseRecord, error) {
	user := config.Instances[instanceName].Users[userName].Username
	host := config.Instances[instanceName].Host
	port := 5432
	dbname := "postgres" // Connect to a default database to list others

	authConfig, err := NewAuthConfig(config.Instances[instanceName].Users[userName].DefaultAuth, config.Instances[instanceName].Users[userName].Auth)
	if err != nil {
		return nil, err
	}

	password := authConfig.GetCredential()
	connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)

	db, err := sql.Open("postgres", connectionUri)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// rows, err := db.QueryContext(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false;")
	rows, err := db.QueryContext(ctx, `SELECT
  d.datname AS "Name",
  pg_catalog.pg_get_userbyid(d.datdba) AS "Owner",
  pg_catalog.pg_encoding_to_char(d.encoding) AS "Encoding",
  d.datcollate AS "Collate",
  d.datctype AS "Ctype",
  pg_catalog.array_to_string(d.datacl, E'\n') AS "Access privileges"
FROM
  pg_catalog.pg_database d
WHERE
  datistemplate = false
ORDER BY
  d.datname;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []DatabaseRecord
	for rows.Next() {
		var name string
		var owner string
		var encoding string
		var collate string
		var ctype string
		var accessPrivileges sql.NullString
		if err := rows.Scan(&name, &owner, &encoding, &collate, &ctype, &accessPrivileges); err != nil {
			return nil, err
		}

		database := DatabaseRecord{
			Name: name,
			Owner: owner,
			Encoding: encoding,
			Collate: collate,
			Ctype: ctype,
		}

		if accessPrivileges.Valid {
			database.AccessPrivileges = accessPrivileges.String
		}
		databases = append(databases, database)
	}

	return databases, nil
}

func getPasswordFromVault(address string, mount_path string, secret_path string, secret_key string) (string, error) {
	config := vault.DefaultConfig()
	config.Address = address

	client, err := vault.NewClient(config)
	if err != nil {
		return "", err
	}

	secret, err := client.KVv2(mount_path).Get(context.Background(), secret_path)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", fmt.Errorf("no secret found at path: %s", secret_path)
	}

	password, ok := secret.Data[secret_key].(string)
	if !ok {
		return "", fmt.Errorf("key '%s' not found in secret data", secret_key)
	}

	return password, nil
}
