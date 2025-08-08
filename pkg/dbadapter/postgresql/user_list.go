package postgresql

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
)

type RoleList struct {
	*PostgreSQLAdapter
}

func NewRoleList(adapter *PostgreSQLAdapter) *RoleList {
	return &RoleList{
		PostgreSQLAdapter: adapter,
	}
}

func (u *RoleList) GetTitle() string {
	return "Roles"
}

func (u *RoleList) GetContent(session *session.Session) tview.Primitive {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	headers := []string{"Role name", "Attributes", "Member of", "Description"}
	for i, header := range headers {
		table.SetCell(0, i, tview.NewTableCell(header).SetSelectable(false))
	}

	go func() {
		session.ShowMessageAsync("Loading roles", false)

		Roles, err := u.listRoles()
		session.CloseMessageAsync()

		if err != nil {
			session.ShowMessageAsync(fmt.Sprintf("Error: %s", err), true)
			return
		}

		session.App.QueueUpdateDraw(func() {
			for i, Role := range Roles {
				table.SetCell(i+1, 0, tview.NewTableCell(Role.RolName))
				table.SetCell(i+1, 1, tview.NewTableCell(Role.Attributes))
				table.SetCell(i+1, 2, tview.NewTableCell(Role.MemberOf))
				table.SetCell(i+1, 3, tview.NewTableCell(Role.Description))
			}
		})
	}()

	return table
}

func (u *RoleList) GetKeyBindings() (keybindings []*session.KeyBinding) {
	return
}

func (u *RoleList) GetInfo() (info []session.Info) {
	info = u.PostgreSQLAdapter.GetInfo()
	return
}
