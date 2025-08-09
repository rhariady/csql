package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rhariady/csql/pkg/auth"
	"github.com/rhariady/csql/pkg/session"
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
		defer func() {
			err := v.ptmx.Close()
			if err != nil {
				s.ShowMessageAsync(fmt.Sprintf("Error closing pty:\n%s", err), true)
			}
		}()
		buf := make([]byte, 4096)
		w := tview.ANSIWriter(terminal)
		for {
			n, err := v.ptmx.Read(buf)
			if err != nil {
				return
			}
			s.App.QueueUpdateDraw(func() {
				_, err = w.Write(buf[:n])
				if err != nil {
					s.ShowMessageAsync(fmt.Sprintf("Error writing to pty:\n%s", err), true)
				}
				terminal.ScrollToEnd()
			})
		}
	}()

	terminal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			_, err := v.ptmx.Write([]byte(string(event.Rune())))
			if err != nil {
				s.ShowMessage(fmt.Sprintf("Error writing to pty:\n%s", err), true)
			}
			return nil
		case tcell.KeyEnter:
			_, err := v.ptmx.Write([]byte{'\n'})
			if err != nil {
				s.ShowMessage(fmt.Sprintf("Error writing to pty:\n%s", err), true)
			}
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
			err := v.cmd.Process.Signal(syscall.SIGINT)
			if err != nil {
				s.ShowMessage(fmt.Sprintf("Error sending signal to psql process:\n%s", err), true)
			}
			return nil
		case tcell.KeyEsc:
			if v.ptmx != nil {
				err := v.ptmx.Close()
				if err != nil {
					s.ShowMessage(fmt.Sprintf("Error closing pty:\n%s", err), true)
				}
			}
			if v.cmd != nil && v.cmd.Process != nil {
				err := v.cmd.Process.Kill()
				if err != nil {
					s.ShowMessage(fmt.Sprintf("Error killing psql process:\n%s", err), true)
				}

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
