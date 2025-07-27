package app

import (
	"fmt"
	
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/dbadapter"
	"github.com/rhariady/csql/pkg/session"
	"github.com/rhariady/csql/pkg/config"
)

type UserList struct{
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

	userTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			session.CloseModal()
		}
	})

	userTable.SetSelectedFunc(func(row int, column int) {
		// databaseList := NewDatabaseList(i.instanceName, userName)
		session.CloseModal()
		userName := userTable.GetCell(row, 0).Text
		user, _ := i.instance.GetUserConfig(userName)
		dbAdapter, _ := dbadapter.GetDBAdapter(i.instance.Type)
		dbAdapter.Connect(session, i.instance, user)
		// session.SetView(databaseList)
		//ShowDatabaseList(app, pages, instanceName, userName, userTable)
	})

	userTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// if event.Rune() == 'a' {
		// 	ShowAddUserForm(app, pages, mainTable, instanceName)
		// 	return nil // Consume the event
		// }
		return event
	})

	return userTable
}

func NewUserList(instance *config.InstanceConfig) *UserList {
	return &UserList{
		instance: instance,
	}
}
