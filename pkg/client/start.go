package client

import (
	"fmt"
	"os"
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
	// Use config session name, or current session name if windows should be
	// created within the current session.
	sessionName := session.Session
	if sessionName == "" {
		return ErrNoSessionName
	}
	if j.Options.Inside {
		var err error
		if sessionName, err = j.Tmux.SessionName(); err != nil {
			return err
		}
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	// Resolve session start directory.
	// If session path is empty, use config path.
	// If session path is "." or "./", use current directory.
	sessionPath := shell.ExpandPath(session.Path)
	if session.Path == "" {
		sessionPath = filepath.Dir(session.ConfigPath)
	}
	if session.Path == "." || session.Path == "./" {
		sessionPath = currentDirectory
	}

	firstWindowName := ""
	if len(session.Windows) > 0 {
		firstWindowName = session.Windows[0].Name
	}

	sessionExists := j.Tmux.SessionExists(sessionName)

	switch {
	case j.Options.Inside:
		// Skip session creation.
	case sessionExists && len(windows) == 0:
		return nil
	case !sessionExists:
		// Execute "before" commands.
		err := j.execShellCommands(session.Before, sessionPath)
		if err != nil {
			return err
		}

		// Create new session and set environment variables.
		_, err = j.Tmux.NewSession(session.Session, sessionPath, firstWindowName)
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

	// Create windows inside the session.
	for i, w := range session.Windows {
		if (len(windows) == 0 && w.Manual) ||
			(len(windows) > 0 && !slices.Contains(windows, w.Name)) {
			continue
		}

		// Resolve window start directory.
		// If window path is empty, use config path.
		// If window path is "." or "./", use current directory.
		windowPath := shell.ExpandPath(w.Path)
		if w.Path == "." || w.Path == "./" {
			windowPath = currentDirectory
		}
		if w.Path == "" || !filepath.IsAbs(w.Path) {
			windowPath = filepath.Join(sessionPath, w.Path)
		}

		// Create the windowID, unless it's the first one.
		windowID := ""
		switch {
		case i > 0 || j.Options.Inside:
			windowID, err = j.Tmux.NewWindow(sessionName, w.Name, windowPath)
			if err != nil {
				return err
			}
		case i == 0 && w.Name != "":
			windowID = w.Name
		case i == 0 && w.Name == "":
			currentWindows, err := j.Tmux.ListWindows(sessionName)
			if err != nil {
				return err
			}
			if len(currentWindows) == 0 {
				return ErrNoWindowsFound
			}
			windowID = currentWindows[0].ID
		}

		// Window already exists? Skipping.
		if windowID == "" {
			continue
		}

		// Optionally focus window.
		if w.Focus {
			err := j.Tmux.SelectWindow(sessionName, windowID)
			if err != nil {
				return err
			}
		}

		// Run commands
		if w.Cmd != "" {
			w.Commands = append(w.Commands, w.Cmd)
		}
		for _, cmd := range w.Commands {
			if session.SuppressHistory {
				cmd = " " + cmd
			}
			time.Sleep(time.Millisecond * time.Duration(session.CommandDelay))
			if err := j.Tmux.SendWindowKeys(sessionName, windowID, cmd); err != nil {
				fmt.Println(err)
			}
		}

		// Create panes
		for _, p := range w.Panes {
			// Resolve pane start directory.
			// If pane path is empty, use config path.
			// If pane path is "." or "./", use current directory.
			panePath := shell.ExpandPath(p.Path)
			if p.Path == "." || p.Path == "./" {
				panePath = currentDirectory
			}
			if p.Path == "" || !filepath.IsAbs(p.Path) {
				panePath = filepath.Join(windowPath, p.Path)
			}

			split := tmux.VSplit
			if p.Type != "" {
				split = tmux.SplitType(p.Type)
			}
			paneID, err := j.Tmux.NewPane(sessionName, windowID, panePath, split)
			if err != nil {
				return err
			}

			// Run commands inside pane.
			if p.Cmd != "" {
				p.Commands = append(p.Commands, p.Cmd)
			}
			for _, cmd := range p.Commands {
				if session.SuppressHistory {
					cmd = " " + cmd
				}
				time.Sleep(time.Millisecond * time.Duration(session.CommandDelay))
				err = j.Tmux.SendPaneKeys(sessionName, windowID, paneID, cmd)
				if err != nil {
					fmt.Println(err)
				}
			}

			// Optionally focus a pane.
			if p.Focus {
				err := j.Tmux.SelectPane(sessionName, windowID, paneID)
				if err != nil {
					return err
				}
			}
		}

		if w.Layout != "" {
			_, err = j.Tmux.SelectLayout(sessionName, windowID, w.Layout)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
