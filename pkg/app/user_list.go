package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/dbadapter"
	"github.com/rhariady/csql/pkg/session"
)

type UserList struct {
	instance *config.InstanceConfig
}

func (i *UserList) GetTitle() string {
	return "Select a user"
}

func (i *UserList) GetContent(session *session.Session) tview.Primitive {
	userTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	userRow := 0
	for _, user := range i.instance.Users {
		userTable.SetCell(userRow, 0, tview.NewTableCell(user.Username))
		userTable.SetCell(userRow, 1, tview.NewTableCell(fmt.Sprintf("[auth=%s]", user.AuthType)).SetExpansion(1))
		userRow++
	}

	userTable.Select(0, 0)

	userTable.SetSelectedFunc(func(row int, column int) {
		// databaseList := NewDatabaseList(i.instanceName, userName)
		session.CloseModal()
		userName := userTable.GetCell(row, 0).Text
		user, err := i.instance.GetUserConfig(userName)
		if err != nil {
			session.ShowMessage(fmt.Sprintf("Error:\n%s", err), true)
		}
		dbAdapter, err := dbadapter.GetDBAdapter(i.instance.Type)
		if err != nil {
			session.ShowMessage(fmt.Sprintf("Error:\n%s", err), true)
			return
		}
		err = dbAdapter.Connect(session, i.instance, user, user.DefaultDatabase)
		if err != nil {
			session.ShowMessage(fmt.Sprintf("Error:\n%s", err), true)
			return
		}

		// session.SetView(databaseList)
		//ShowDatabaseList(app, pages, instanceName, userName, userTable)
	})

	userTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'a':
			session.CloseModal()
			add_user := NewAddUser(i.instance)
			session.ShowModal(add_user)
			return nil
		}
		return event
	})

	return userTable
}

func (i *UserList) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("(a)", "Add new user"),
		session.NewKeyBinding("<enter>", "Select user"),
	}
	return
}

func (i *UserList) GetInfo() (info []session.Info) {
	return
}

func NewUserList(instance *config.InstanceConfig) *UserList {
	return &UserList{
		instance: instance,
	}
}

func (i *UserList) ExecuteCommand(s *session.Session, command string) error {
	return nil
}
