
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

type ShellView struct {
	*PostgreSQLAdapter
	ptmx *os.File
	cmd  *exec.Cmd
}

func NewShellView(adapter *PostgreSQLAdapter) *ShellView {
	return &ShellView{
		PostgreSQLAdapter: adapter,
	}
}

func (v *ShellView) GetTitle() string {
	return "Shell"
}

func (v *ShellView) GetContent(s *session.Session) tview.Primitive {
	terminal := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			s.App.Draw()
		})

	terminal.SetBorder(false)
	authConfig, err := auth.GetAuth(v.user.AuthType, v.user.AuthParams)
	if err != nil {
		s.ShowMessage(fmt.Sprintf("Error:\n%s", err), true)
	}

	password, err := authConfig.GetCredential()

	if err != nil {
		s.ShowMessage(fmt.Sprintf("Error:\n%s", err), true)
	}

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		v.user.Username,
		password,
		v.instance.Host,
		v.instance.Port,
		v.database,
	)

	v.cmd = exec.Command("psql", dsn)
	v.ptmx, err = pty.Start(v.cmd)
	if err != nil {
		s.ShowMessage(fmt.Sprintf("Error starting shell: %s", err), true)
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

func (v *ShellView) GetKeyBindings() (keybindings []*session.KeyBinding) {
	keybindings = []*session.KeyBinding{
		session.NewKeyBinding("<escape>", "Go back to table list"),
	}

	return
}

func (v *ShellView) GetInfo() (info []session.Info) {
	info = v.PostgreSQLAdapter.GetInfo()
	return
}
