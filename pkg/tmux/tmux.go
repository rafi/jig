package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rafi/jig/pkg/shell"
)

type TmuxClient struct {
	Bin string
	Cmd shell.Commander
}

const ColumnSep = "§"

// NewSession creates a new session with optional name and directory.
func (t TmuxClient) NewSession(name, dir, windowName string) (string, error) {
	args := []string{"new-session", "-Pd", "-F", "#{session_id}"}
	if name != "" {
		args = append(args, "-s", name)
	}
	// Naming a window will disable automatic-rename.
	if windowName != "" {
		args = append(args, "-n", windowName)
	}
	if dir != "" {
		args = append(args, "-c", shell.ExpandPath(dir))
	}
	return t.Cmd.Exec(exec.Command(t.Bin, args...))
}

// NewWindow creates a new window with optional name and directory.
func (t TmuxClient) NewWindow(target Target, name, dir string) (string, error) {
	args := []string{"new-window", "-Pd", "-t", target.Get()}

	// Naming a window will disable automatic-rename.
	if name != "" {
		args = append(args, "-n", name)
	}
	args = append(args, "-F", "#{window_id}")
	if dir != "" {
		args = append(args, "-c", shell.ExpandPath(dir))
	}

	cmd := exec.Command(t.Bin, args...)
	return t.Cmd.Exec(cmd)
}

// NewPane creates a new split in a session's window.
func (t TmuxClient) NewPane(target Target, dir, split string) (string, error) {
	args := []string{"split-window", "-Pd", "-t", target.Get()}

	switch split {
	case "v", "-v", "vertical":
		args = append(args, "-v")
	case "h", "-h", "horizontal":
		args = append(args, "-h")
	default:
		fmt.Printf("Invalid split type: %s\n", split)
	}

	if dir != "" {
		args = append(args, "-c", shell.ExpandPath(dir))
	}
	args = append(args, "-F", "#{pane_id}")

	cmd := exec.Command(t.Bin, args...)
	return t.Cmd.Exec(cmd)
}

// KillWindow kills a window in a session.
func (t TmuxClient) KillWindow(target Target) error {
	cmd := exec.Command(t.Bin, "kill-window", "-t", target.Get())
	_, err := t.Cmd.Exec(cmd)
	return err
}

// SendKeys sends key-strokes to a target.
func (t TmuxClient) SendKeys(target Target, command string) error {
	baseArgs := []string{"send-keys", "-t", target.Get()}
	cmd := exec.Command(t.Bin, append(baseArgs, "-l", command)...)
	err := t.Cmd.ExecSilently(cmd)
	_ = t.Cmd.ExecSilently(exec.Command(t.Bin, append(baseArgs, "Enter")...))
	return err
}

// Attach attaches to a session.
func (t TmuxClient) Attach(
	session string,
	stdin, stdout, stderr *os.File,
) error {
	cmd := exec.Command(t.Bin, "attach", "-d", "-t", session)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return t.Cmd.ExecSilently(cmd)
}

// SwitchClient switches to a client.
func (t TmuxClient) SwitchClient(session string) error {
	cmd := exec.Command(t.Bin, "switch-client", "-t", session)
	return t.Cmd.ExecSilently(cmd)
}

// SessionExists checks if a session exists.
func (t TmuxClient) SessionExists(name string) bool {
	cmd := exec.Command(t.Bin, "has-session", "-t", name+":")
	res, err := t.Cmd.Exec(cmd)
	return res == "" && err == nil
}

// SessionName returns the current session name.
func (t TmuxClient) SessionName() (string, error) {
	cmd := exec.Command(t.Bin, "display-message", "-p", "#S")
	return t.Cmd.Exec(cmd)
}

// SetEnv sets an environment variable in a session.
func (t TmuxClient) SetEnv(session, key, value string) (string, error) {
	cmd := exec.Command(t.Bin, "setenv", "-t", session, key, value)
	return t.Cmd.Exec(cmd)
}

// RenumberWindows renumbers windows' index in a session.
func (t TmuxClient) RenumberWindows(session string) error {
	cmd := exec.Command(t.Bin, "move-window", "-r", "-s", session, "-t", session)
	return t.Cmd.ExecSilently(cmd)
}

// SelectLayout selects a layout for a window.
func (t TmuxClient) SelectLayout(target Target, layout string) (string, error) {
	cmd := exec.Command(t.Bin, "select-layout", "-t", target.Get(), layout)
	return t.Cmd.Exec(cmd)
}

// SelectWindow selects a window in a session.
func (t TmuxClient) SelectWindow(target Target) error {
	cmd := exec.Command(t.Bin, "select-window", "-t", target.Get())
	return t.Cmd.ExecSilently(cmd)
}

// SelectPane selects a pane in a window.
func (t TmuxClient) SelectPane(target Target) error {
	cmd := exec.Command(t.Bin, "select-pane", "-t", target.Get())
	return t.Cmd.ExecSilently(cmd)
}

// StopSession stops a session.
func (t TmuxClient) StopSession(target Target) (string, error) {
	cmd := exec.Command(t.Bin, "kill-session", "-t", target.Get())
	return t.Cmd.Exec(cmd)
}

// ListSessions returns a list of sessions and their information.
func (t TmuxClient) ListSessions() ([]TmuxSession, error) {
	fields := getFormat(TmuxSession{})
	format := strings.Join(fields, ColumnSep)
	cmd := exec.Command(t.Bin, "list-sessions", "-F", format)
	out, err := t.Cmd.Exec(cmd)
	if err != nil {
		return []TmuxSession{}, err
	}

	lines := strings.Split(out, "\n")
	sessions := make([]TmuxSession, len(lines))
	for i, line := range lines {
		session := TmuxSession{}
		if err := parseOutput(line, &session); err != nil {
			return nil, err
		}
		sessions[i] = session
	}
	return sessions, nil
}

// ListWindows returns a list of windows and their information.
func (t TmuxClient) ListWindows(target Target) ([]TmuxWindow, error) {
	fields := getFormat(TmuxWindow{})
	format := strings.Join(fields, ColumnSep)
	cmd := exec.Command(t.Bin, "list-windows", "-t", target.Get(), "-F", format)
	out, err := t.Cmd.Exec(cmd)
	if err != nil {
		return []TmuxWindow{}, err
	}

	lines := strings.Split(out, "\n")
	windows := make([]TmuxWindow, len(lines))
	for i, line := range lines {
		window := TmuxWindow{}
		if err := parseOutput(line, &window); err != nil {
			return nil, err
		}
		windows[i] = window
	}
	return windows, nil
}

// ListPanes returns a list of panes in a window.
func (t TmuxClient) ListPanes(target Target) ([]TmuxPane, error) {
	fields := getFormat(TmuxPane{})
	format := strings.Join(fields, ColumnSep)
	cmd := exec.Command(t.Bin, "list-panes", "-t", target.Get(), "-F", format)
	out, err := t.Cmd.Exec(cmd)
	if err != nil {
		return []TmuxPane{}, err
	}

	lines := strings.Split(out, "\n")
	panes := make([]TmuxPane, len(lines))
	for i, line := range lines {
		pane := TmuxPane{}
		if err := parseOutput(line, &pane); err != nil {
			return nil, err
		}
		panes[i] = pane
	}
	return panes, nil
}
