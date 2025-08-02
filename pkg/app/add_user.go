package app

import (	
	_ "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/auth"
	"github.com/rhariady/csql/pkg/session"
)

type AddUser struct {
	instance *config.InstanceConfig
}

func (a *AddUser) GetTitle() string {
	return "Add New User"
}

// func ShowAddUserForm(app *tview.Application, pages *tview.Pages, mainTable *tview.Table, instanceName string) {
func (a *AddUser) GetContent(s *session.Session) tview.Primitive {
	var form *tview.Form
	auth_type := tview.NewDropDown().
		SetLabel("Auth Type")
		// SetListStyles(tcell.StyleDefault.Background(tcell.ColorNone), tcell.StyleDefault.Background(tcell.ColorGrey)).
		// SetFocusedStyle(tcell.StyleDefault.Background(tcell.ColorGrey)).
		// SetPrefixStyle(tcell.StyleDefault.Background(tcell.ColorGrey))

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

			a.instance = s.Config.AddInstanceUser(a.instance.Name, newUser)
			s.Config.WriteConfig()

			s.CloseModal()

			user_list := NewUserList(a.instance)
			s.ShowModal(user_list)
			// pages.RemovePage("addUserModal")
			// pages.RemovePage("userSelectionModal")
			// ShowUserSelection(app, pages, mainTable, instanceName)
		}).
		AddButton("Cancel", func() {
			s.CloseModal()
			// pages.RemovePage("addUserModal")
		})

	// form.SetBorder(true).SetTitle("Add New User")
	// form.SetFieldBackgroundColor(tcell.ColorDarkGreen)
	// fieldStyle := tcell.StyleDefault.
	// 	Background(tcell.ColorGrey).
	// 	Blink(true).
	// 	Underline(tcell.ColorWhite)
	// form.SetFieldStyle(fieldStyle)
	// form.SetLabelColor(tcell.ColorDarkGreen)
	// form.SetTitleColor(tcell.ColorDarkGreen)

	auth_type.SetCurrentOption(0)

	// modal := tview.NewFlex().
	// 	AddItem(nil, 0, 1, false).
	// 	AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
	// 		AddItem(nil, 0, 1, false).
	// 		AddItem(form, 0, 1, true).
	// 		AddItem(nil, 0, 1, false), 0, 1, false).
	// 	AddItem(nil, 0, 1, false)

	// pages.AddPage("addUserModal", modal, true, true)
	// app.SetFocus(form)

	return form
}

func (a *AddUser) GetKeyBindings() (keybindings []*session.KeyBinding) {
	return
}

func NewAddUser(instance *config.InstanceConfig) *AddUser {
	return &AddUser{instance}
}
