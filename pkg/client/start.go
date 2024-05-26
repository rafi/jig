package client

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/rafi/jig/pkg/shell"
	"github.com/rafi/jig/pkg/tmux"
)

// Start starts a new tmux session, any nested sessions, run optional `before`
// command and optionally attach to the first session.
func (j Jig) Start(config Config, windows []string) error {
	if j.Options.Inside && !j.InSession {
		return ErrNotInsideSession
	}

	for _, s := range config.Sessions {
		if err := j.startSession(s, windows); err != nil {
			return err
		}
	}
	if err := j.startSession(config, windows); err != nil {
		return err
	}

	// Attach/switch to the session.
	if j.Options.Detach || j.Options.Inside {
		return nil
	}
	return j.SwitchOrAttach(config.Session)
}

// startSession starts a new tmux session, creates all windows and panes.
func (j Jig) startSession(session Config, windows []string) error {
	var err error

	// Use config session name, or current session name if windows should be
	// created within the current session.
	sessionName := session.Session
	if sessionName == "" {
		return ErrNoSessionName
	}
	if j.Options.Inside {
		if sessionName, err = j.Tmux.SessionName(); err != nil {
			return err
		}
	}

	// Resolve session start directory.
	if session.Path, err = session.GetSessionPath(); err != nil {
		return err
	}

	firstWinName := ""
	if len(session.Windows) > 0 {
		firstWinName = session.Windows[0].Name
	}

	sessionExists := j.Tmux.SessionExists(sessionName)

	switch {
	case j.Options.Inside:
		// Skip session creation.
	case sessionExists && len(windows) == 0:
		return nil
	case !sessionExists:
		// Execute "before" commands.
		err := j.execShellCommands(session.Before, session.Path)
		if err != nil {
			return err
		}

		// Create new session and set environment variables.
		_, err = j.Tmux.NewSession(session.Session, session.Path, firstWinName)
		if err != nil {
			return err
		}
		if len(session.Env) > 0 {
			err = j.setEnvVariables(session.Session, session.Env)
			if err != nil {
				return err
			}
		}
	}
	return j.createSessionWindows(session, windows)
}

// createSessionWindows creates windows inside the session.
func (j Jig) createSessionWindows(session Config, explicitWindows []string) error {
	var err error
	target := tmux.Target{Session: session.Session}
	for i, w := range session.Windows {
		if (len(explicitWindows) == 0 && w.Manual) ||
			(len(explicitWindows) > 0 && !slices.Contains(explicitWindows, w.Name)) {
			continue
		}

		// Resolve window start directory.
		if w.Path != "" {
			w.Path = shell.ExpandPath(w.Path)
		}
		if w.Path == "" || !filepath.IsAbs(w.Path) {
			w.Path = filepath.Join(session.Path, w.Path)
		}

		// Create a window, unless it's the first one.
		target.Window = ""
		target.Pane = ""
		switch {
		case i > 0 || j.Options.Inside:
			target.Window, err = j.Tmux.NewWindow(target, w.Name, w.Path)
			if err != nil {
				return err
			}

		// If processing 1st window, and it's named - then use its name as id.
		case i == 0 && w.Name != "":
			target.Window = w.Name

		// If first window is unnamed, ask tmux for the session's first window.
		case i == 0 && w.Name == "":
			currentWindows, err := j.Tmux.ListWindows(target)
			if err != nil {
				return err
			}
			if len(currentWindows) == 0 {
				return ErrNoWindowsFound
			}
			target.Window = currentWindows[0].ID
		}

		// Optionally focus window.
		if w.Focus {
			err := j.Tmux.SelectWindow(target)
			if err != nil {
				return err
			}
		}

		// Run window commands.
		newWinCommands := w.GetCommands()
		for _, cmd := range newWinCommands {
			if session.SuppressHistory {
				cmd = " " + cmd
			}
			time.Sleep(time.Millisecond * time.Duration(session.CommandDelay))
			err := j.Tmux.SendKeys(target, cmd)
			if err != nil {
				fmt.Println(err)
			}
		}

		// Create panes.
		for _, p := range w.Panes {
			// Resolve pane start directory.
			if p.Path != "" {
				p.Path = shell.ExpandPath(p.Path)
			}
			panePath := shell.ExpandPath(p.Path)
			if p.Path == "" || !filepath.IsAbs(panePath) {
				panePath = filepath.Join(w.Path, p.Path)
			}

			target.Pane, err = j.Tmux.NewPane(target, panePath, p.Type)
			if err != nil {
				return err
			}

			// Run commands inside pane.
			for _, cmd := range p.GetCommands() {
				if session.SuppressHistory {
					cmd = " " + cmd
				}
				time.Sleep(time.Millisecond * time.Duration(session.CommandDelay))
				err := j.Tmux.SendKeys(target, cmd)
				if err != nil {
					fmt.Println(err)
				}
			}

			// Optionally focus a pane.
			if p.Focus {
				if err := j.Tmux.SelectPane(target); err != nil {
					return err
				}
			}
		}

		target.Pane = ""
		if w.Layout != "" {
			_, err := j.Tmux.SelectLayout(target, w.Layout)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
