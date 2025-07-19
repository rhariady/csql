package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-isatty"
	"github.com/rivo/tview"
	vault "github.com/hashicorp/vault/api"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

var config *Config

func main() {
	CheckConfigFile()
	var err error
	config, err = GetConfig()
	if err != nil {
		panic(err)
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		app := tview.NewApplication()

		// Create the main database table
		table := tview.NewTable().
			SetBorders(false).
			SetSelectable(true, false)

		// Populate the table with instances
		table.SetCell(0, 0, tview.NewTableCell("Name").SetSelectable(false)).
			SetCell(0, 1, tview.NewTableCell("Project ID").SetSelectable(false)).
			SetCell(0, 2, tview.NewTableCell("Host").SetSelectable(false))

		table.SetWrapSelection(true, true)

		row := 1
		for name, instance := range config.Instances {
			table.SetCell(row, 0, tview.NewTableCell(name))
			table.SetCell(row, 1, tview.NewTableCell(instance.ProjectID))
			table.SetCell(row, 2, tview.NewTableCell(instance.Host))
			row++
		}

		// Set the selected function for the table (triggered by Enter key)
		table.SetSelectedFunc(func(row int, column int) {
			if row == 0 { // Skip header row
				return
			}
			instanceName := table.GetCell(row, 0).Text
			showUserSelection(app, table, instanceName)
		})

		// Set input capture for 'a' key to trigger the same selection logic
		table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Rune() == 'a' {
				discoverDatabases()
				return nil // Consume the event
			}
			return event
		})

		flex := tview.NewFlex().AddItem(table, 0, 1, true)
		flex.SetBorder(true).SetTitle("Databases")

		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			}
			return event
		})

		if err := app.SetRoot(flex, true).Run(); err != nil {
			panic(err)
		}
	} else {
		fmt.Println("This application is intended to be run in an interactive terminal.")
	}
}

func showUserSelection(app *tview.Application, mainTable *tview.Table, instanceName string) {
	userTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	userTable.SetCell(0, 0, tview.NewTableCell("Username")).
		SetCell(0, 1, tview.NewTableCell("Auth Type"))

	userRow := 1
	for name, user := range config.Instances[instanceName].Users {
		userTable.SetCell(userRow, 0, tview.NewTableCell(name))
		userTable.SetCell(userRow, 1, tview.NewTableCell(user.DefaultAuth))
		userRow++
	}

	userTable.Select(1, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.SetRoot(mainTable, true) // Go back to instance table
		}
	}).SetSelectedFunc(func(row int, column int) {
		userName := userTable.GetCell(row, 0).Text
		app.Stop()
		user := config.Instances[instanceName].Users[userName].Username
		host := config.Instances[instanceName].Host
		port := 5432
		dbname := "postgres"

		authConfig, err := NewAuthConfig(config.Instances[instanceName].Users[userName].DefaultAuth, config.Instances[instanceName].Users[userName].Auth)
		if err != nil {
			panic(err)
		}

		password := authConfig.GetCredential()
		connectionUri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", user, password, host, port, dbname)

		fmt.Println("Connecting to:", connectionUri)
		cmd := exec.Command("psql", connectionUri)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Error:", err)
		}
	})

	userTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'a' {
			showAddUserForm(app, mainTable, instanceName)
			return nil // Consume the event
		}
		return event
	})

	app.SetRoot(userTable, true)
}

func showAddUserForm(app *tview.Application, mainTable *tview.Table, instanceName string) {
	var form *tview.Form
	form = tview.NewForm().
		AddInputField("Username", "", 20, nil, nil).
		AddDropDown("Auth Type", []string{"vault"}, 0, nil).
		AddInputField("Vault Address", "", 40, nil, nil).
		AddInputField("Vault Mount Path", "", 40, nil, nil).
		AddInputField("Vault Secret Path", "", 40, nil, nil).
		AddInputField("Vault Secret Key", "", 40, nil, nil).
		AddButton("Add User", func() {
			username := form.GetFormItem(0).(*tview.InputField).GetText()
			_, authType := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
			vaultAddress := form.GetFormItem(2).(*tview.InputField).GetText()
			vaultMountPath := form.GetFormItem(3).(*tview.InputField).GetText()
			vaultSecretPath := form.GetFormItem(4).(*tview.InputField).GetText()
			vaultSecretKey := form.GetFormItem(5).(*tview.InputField).GetText()

			if authType == "vault" {
				newUser := UserConfig{
					Username: username,
					DefaultAuth: "vault",
					Auth: map[string]interface{}{
						"vault": map[string]string{
							"address": vaultAddress,
							"mount_path": vaultMountPath,
							"secret_path": vaultSecretPath,
							"secret_key": vaultSecretKey,
						},
					},
				}
				config.Instances[instanceName].Users[username] = newUser
				config.WriteConfig()
			}
			showUserSelection(app, mainTable, instanceName)
		}).
		AddButton("Cancel", func() {
			showUserSelection(app, mainTable, instanceName)
		})
	app.SetRoot(form, true)
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
