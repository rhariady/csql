package app

import (
	"fmt"

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
			for form.GetFormItemCount() > 3 {
				form.RemoveFormItem(3)
			}
			authConfig.GetFormInput(form)
		})
	}

	form = tview.NewForm().
		AddInputField("Username", "", 0, nil, nil).
		AddInputField("Default database", "", 0, nil, nil).
		AddFormItem(auth_type).
		AddButton("Add User", func() {
			username := form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
			default_database := form.GetFormItemByLabel("Default database").(*tview.InputField).GetText()
			_, authType := form.GetFormItem(2).(*tview.DropDown).GetCurrentOption()

			authAdapter, err := auth.GetAuth(authType, nil)
			if err != nil {
				s.ShowMessage(fmt.Sprintf("Error:\n%s", err), true)
			}

			authParams := authAdapter.ParseFormInput(form)
			newUser := config.UserConfig{
				Username:   username,
				DefaultDatabase: default_database,
				AuthType:   authType,
				AuthParams: authParams,
			}

			a.instance = s.Config.AddInstanceUser(a.instance.Name, newUser)
			s.Config.WriteConfig()

			s.CloseModal()

			user_list := NewUserList(a.instance)
			s.ShowModal(user_list)
		}).
		AddButton("Cancel", func() {
			s.CloseModal()
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

	return form
}

func (a *AddUser) GetKeyBindings() (keybindings []*session.KeyBinding) {
	return
}

func (i *AddUser) GetInfo() (info []session.Info) {
	return
}

func (i *AddUser) ExecuteCommand(s *session.Session, command string) error {
	return nil
}

func NewAddUser(instance *config.InstanceConfig) *AddUser {
	return &AddUser{instance}
}
