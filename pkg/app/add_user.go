package app

import (	
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/config"
	"github.com/rhariady/csql/pkg/auth"
)

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
			// ShowUserSelection(app, pages, mainTable, instanceName)
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
