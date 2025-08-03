
package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/session"
	"github.com/rhariady/csql/pkg/auth"
)

type PsqlView struct {
	*PostgreSQLAdapter
	ptmx *os.File
	cmd  *exec.Cmd
}

func NewPsqlView(adapter *PostgreSQLAdapter) *PsqlView {
	return &PsqlView{
		PostgreSQLAdapter: adapter,
	}
}

func (v *PsqlView) GetTitle() string {
	return "psql"
}

func (v *PsqlView) GetContent(s *session.Session) tview.Primitive {
	terminal := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			s.App.Draw()
		})

	terminal.SetBorder(true).SetTitle("psql")
	authConfig, _ := auth.GetAuth(v.user.AuthType, v.user.AuthParams)

	password := authConfig.GetCredential()

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		v.user.Username,
		password,
		v.instance.Host,
		v.instance.Port,
		v.database,
	)

	var err error
	
	v.cmd = exec.Command("psql", dsn)
	v.ptmx, err = pty.Start(v.cmd)
	if err != nil {
		s.App.QueueUpdateDraw(func() {
			s.ShowMessageAsync(fmt.Sprintf("Error starting psql: %s", err), true)
		})
		return terminal
	}

	go func() {
		defer v.ptmx.Close()
		buf := make([]byte, 4096)
		w := tview.ANSIWriter(terminal)
		for {
			n, err := v.ptmx.Read(buf)
			if err != nil {
				return
			}
			s.App.QueueUpdateDraw(func() {
				w.Write(buf[:n])
				terminal.ScrollToEnd()
			})
		}
	}()

	terminal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			v.ptmx.Write([]byte(string(event.Rune())))
			return nil
		case tcell.KeyEnter:
			v.ptmx.Write([]byte{'\n'})
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			currentText := terminal.GetText(false) // false to get raw text without regions
			if len(currentText) > 0 {
				newText := currentText[:len(currentText)-1]
				terminal.SetText(newText)
				// terminal.SetText(newText)
				// s.ShowMessage(newText, true)
			}
			return nil
		case tcell.KeyCtrlC:
			v.cmd.Process.Signal(syscall.SIGINT)
			return nil
		case tcell.KeyEsc:
			if v.ptmx != nil {
				v.ptmx.Close()
			}
			if v.cmd != nil && v.cmd.Process != nil {
				v.cmd.Process.Kill()
			}
			tableList := NewTableList(v.PostgreSQLAdapter)
			s.SetView(tableList)
			return nil
		}
		return event
	})

	return terminal
}

func (v *PsqlView) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("<escape>", "Go back to table list"),
	}
	return
}

func (v *PsqlView) GetInfo() (info []session.Info) {
	info = v.PostgreSQLAdapter.GetInfo()
	return
}
