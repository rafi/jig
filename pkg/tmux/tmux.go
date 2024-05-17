package tmux

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

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
	if windowName != "" {
		args = append(args, "-n", windowName)
	}
	if dir != "" {
		args = append(args, "-c", shell.ExpandPath(dir))
	}
	return t.Cmd.Exec(exec.Command(t.Bin, args...))
}

// NewWindow creates a new window with optional name and directory.
func (t TmuxClient) NewWindow(session, name, dir string) (string, error) {
	args := []string{
		"new-window", "-Pda",
		"-t", session + ":",
		"-F", "#{window_id}",
	}
	if dir != "" {
		args = append(args, "-c", shell.ExpandPath(dir))
	}
	// Naming a window will disable automatic-rename.
	if name != "" {
		args = append(args, "-n", name)
	}
	cmd := exec.Command(t.Bin, args...)
	return t.Cmd.Exec(cmd)
}

// NewPane splits a window in a session.
func (t TmuxClient) NewPane(session, window, dir string, split SplitType,
) (string, error) {
	args := []string{
		"split-window", "-Pd",
		"-t", session + ":" + window,
	}

	switch split {
	case VSplit:
		fallthrough
	case HSplit:
		args = append(args, string(split))
	}

	if dir != "" {
		args = append(args, "-c", shell.ExpandPath(dir))
	}
	args = append(args, "-F", "#{pane_id}")

	cmd := exec.Command(t.Bin, args...)
	return t.Cmd.Exec(cmd)
}

// SessionExists checks if a session exists.
func (t TmuxClient) SessionExists(name string) bool {
	cmd := exec.Command(t.Bin, "has-session", "-t", name+":")
	res, err := t.Cmd.Exec(cmd)
	return res == "" && err == nil
}

// KillWindow kills a window in a session.
func (t TmuxClient) KillWindow(session, window string) error {
	cmd := exec.Command(t.Bin, "kill-window", "-t", session+":"+window)
	_, err := t.Cmd.Exec(cmd)
	return err
}

// SendKeys sends key-strokes to a specific window.
func (t TmuxClient) SendWindowKeys(session, window, command string) error {
	return t.sendKeys(session+":"+window, command)
}

// SendKeys sends key-strokes to a specific window.
func (t TmuxClient) SendPaneKeys(session, window, pane, command string) error {
	return t.sendKeys(session+":"+window+"."+pane, command)
}

// sendKeys sends key-strokes to a flexible target.
func (t TmuxClient) sendKeys(target, command string) error {
	baseArgs := []string{"send-keys", "-t", target}
	cmd := exec.Command(t.Bin, append(baseArgs, "-l", command)...)
	err := t.Cmd.ExecSilently(cmd)
	_ = t.Cmd.ExecSilently(exec.Command(t.Bin, append(baseArgs, "Enter")...))
	return err
}

// Attach attaches to a session.
func (t TmuxClient) Attach(
	target string,
	stdin, stdout, stderr *os.File,
) error {
	cmd := exec.Command(t.Bin, "attach", "-d", "-t", target)

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return t.Cmd.ExecSilently(cmd)
}

// RenumberWindows renumbers windows' index in a session.
func (t TmuxClient) RenumberWindows(session string) error {
	cmd := exec.Command(t.Bin, "move-window", "-r", "-s", session, "-t", session)
	return t.Cmd.ExecSilently(cmd)
}

// SelectLayout selects a layout for a window.
func (t TmuxClient) SelectLayout(session, window, layout string,
) (string, error) {
	cmd := exec.Command(t.Bin, "select-layout", "-t", session+":"+window, layout)
	return t.Cmd.Exec(cmd)
}

// SelectWindow selects a window in a session.
func (t TmuxClient) SelectWindow(session, window string) error {
	cmd := exec.Command(t.Bin, "select-window", "-t", session+":"+window)
	return t.Cmd.ExecSilently(cmd)
}

// SelectPane selects a pane in a window.
func (t TmuxClient) SelectPane(session, window, pane string) error {
	cmd := exec.Command(t.Bin, "select-pane", "-t", session+":"+window+"."+pane)
	return t.Cmd.ExecSilently(cmd)
}

// SetEnv sets an environment variable in a session.
func (t TmuxClient) SetEnv(target, key, value string) (string, error) {
	cmd := exec.Command(t.Bin, "setenv", "-t", target, key, value)
	return t.Cmd.Exec(cmd)
}

// StopSession stops a session.
func (t TmuxClient) StopSession(target string) (string, error) {
	cmd := exec.Command(t.Bin, "kill-session", "-t", target)
	return t.Cmd.Exec(cmd)
}

// SwitchClient switches to a client.
func (t TmuxClient) SwitchClient(target string) error {
	cmd := exec.Command(t.Bin, "switch-client", "-t", target)
	return t.Cmd.ExecSilently(cmd)
}

// SessionName returns the current session name.
func (t TmuxClient) SessionName() (string, error) {
	cmd := exec.Command(t.Bin, "display-message", "-p", "#S")
	return t.Cmd.Exec(cmd)
}

// ListSessions returns a list of sessions and their information.
func (t TmuxClient) ListSessions() ([]TmuxSession, error) {
	format := []string{
		"#{session_id}",
		"#{session_attached}",
		"#{session_name}",
		"#{session_marked}",
		"#{session_windows}",
		"#{session_stack}",
		"#{session_created}",
		"#{session_activity}",
		"#{session_last_attached}",
		"#{session_path}",
		"#{session_alerts}",
	}

	cmd := exec.Command(
		t.Bin,
		"list-sessions",
		"-F", strings.Join(format, ColumnSep),
	)
	out, err := t.Cmd.Exec(cmd)
	if err != nil {
		return []TmuxSession{}, err
	}

	lines := strings.Split(out, "\n")
	sessions := make([]TmuxSession, len(lines))
	for i, line := range lines {
		columns := strings.Split(line, ColumnSep)
		if len(columns) != len(format) {
			return sessions, ErrInvalidFormat
		}
		session := TmuxSession{
			ID:       columns[0],
			Attached: columns[1] == "1",
			Name:     columns[2],
			Marked:   columns[3] == "1",
			Stack:    columns[5],
			Path:     columns[9],
			Alerts:   columns[10],
		}

		session.Windows, err = strconv.Atoi(columns[4])
		if err != nil {
			return sessions, err
		}
		session.Created, err = ParseUnixTime(columns[6])
		if err != nil {
			return sessions, err
		}
		session.Activity, err = ParseUnixTime(columns[7])
		if err != nil {
			return sessions, err
		}
		session.LastAttached, err = ParseUnixTime(columns[8])
		if err != nil {
			return sessions, err
		}
		sessions[i] = session
	}
	return sessions, nil
}

// ListWindows returns a list of windows and their information.
func (t TmuxClient) ListWindows(target string) ([]TmuxWindow, error) {
	format := []string{
		"#{window_id}",
		"#{window_name}",
		"#{window_layout}",
		"#{pane_current_path}",
	}

	cmd := exec.Command(
		t.Bin,
		"list-windows",
		"-t", target,
		"-F", strings.Join(format, ColumnSep),
	)
	out, err := t.Cmd.Exec(cmd)
	if err != nil {
		return []TmuxWindow{}, err
	}

	lines := strings.Split(out, "\n")
	windows := make([]TmuxWindow, len(lines))
	for i, w := range lines {
		columns := strings.Split(w, ColumnSep)
		if len(columns) != len(format) {
			return windows, ErrInvalidFormat
		}
		window := TmuxWindow{
			ID:     columns[0],
			Name:   columns[1],
			Layout: columns[2],
			Path:   columns[3],
		}
		windows[i] = window
	}
	return windows, nil
}

// ListPanes returns a list of panes in a window.
func (t TmuxClient) ListPanes(session, window string) ([]TmuxPane, error) {
	format := []string{"#{pane_current_path}", "#{pane_current_command}"}
	cmd := exec.Command(
		t.Bin,
		"list-panes",
		"-t", session+":"+window,
		"-F", strings.Join(format, ColumnSep),
	)
	out, err := t.Cmd.Exec(cmd)
	if err != nil {
		return []TmuxPane{}, err
	}

	lines := strings.Split(out, "\n")
	panes := make([]TmuxPane, len(lines))
	for i, p := range lines {
		columns := strings.Split(p, ColumnSep)
		if len(columns) != len(format) {
			return panes, ErrInvalidFormat
		}
		pane := TmuxPane{
			Path:    columns[0],
			Command: columns[1],
		}
		panes[i] = pane
	}
	return panes, nil
}

func ParseUnixTime(epoch string) (time.Time, error) {
	createdEpoch, err := strconv.ParseInt(epoch, 10, 0)
	if err != nil {
		return time.Time{}, err
	}
	created := time.Unix(createdEpoch, 0)
	return created, nil
}
