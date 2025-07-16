package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

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
		list := tview.NewList().
			AddItem("Connect to a database", "", 'a', func() {
				app.Stop()
				connectToDatabase()
			}).
			AddItem("Discover databases", "", 'b', func() {
				app.Stop()
				discoverDatabases()
			}).
			AddItem("Quit", "", 'q', func() {
				app.Stop()
			})

		if err := app.SetRoot(list, true).Run(); err != nil {
			panic(err)
		}
	} else {
		fmt.Println("This application is intended to be run in an interactive terminal.")
	}
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

func connectToDatabase() {
	app := tview.NewApplication()

	var instanceNames []string
	for name := range config.Instances {
		instanceNames = append(instanceNames, name)
	}

	list := tview.NewList()

	for _, instanceName := range instanceNames {
		list.AddItem(instanceName, "", 0, func() {
			var userNames []string
			for name := range config.Instances[instanceName].Users {
				userNames = append(userNames, name)
			}

			userList := tview.NewList()
			for _, userName := range userNames {
				userList.AddItem(userName, "", 0, func() {
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
			}
			app.SetRoot(userList, true)
		})
	}

	if err := app.SetRoot(list, true).Run(); err != nil {
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
