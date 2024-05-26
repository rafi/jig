package client

import (
	"os"
	"path/filepath"

	"github.com/rafi/jig/pkg/tmux"
)

// GenerateSessionConfig creates a Config object from a tmux session.
func (j Jig) GenerateSessionConfig(sessionName string) (Config, error) {
	var err error
	config := Config{}

	config.Session, err = j.Tmux.SessionName()
	if err != nil {
		return config, err
	}

	target := tmux.Target{Session: sessionName}
	tmuxWindows, err := j.Tmux.ListWindows(target)
	if err != nil {
		return config, err
	}

	currentShell := filepath.Base(os.Getenv("SHELL"))

	for _, w := range tmuxWindows {
		target.Window = w.ID
		tmuxPanes, err := j.Tmux.ListPanes(target)
		if err != nil {
			return config, err
		}

		window := Window{
			Name:   w.Name,
			Layout: w.Layout,
			Panes:  []Pane{},
		}

		// Set session's path to the first found window's path.
		if config.Path == "" {
			config.Path = w.Path
		}

		// Skip window path if it is identical to session's path.
		if w.Path != config.Path {
			window.Path = w.Path
		}

		for _, tmuxPane := range tmuxPanes {
			pane := Pane{Path: tmuxPane.Path}
			if tmuxPane.Command != currentShell {
				pane.Cmd = tmuxPane.Command
			}
			// Skip pane path if it is identical to window or session path.
			if pane.Path == w.Path || (w.Path == "" && pane.Path == config.Path) {
				pane.Path = ""
			}
			// Do not create a pane collection if there's only a single one.
			if len(tmuxPanes) == 1 && pane.Path == "" {
				window.Cmd = pane.Cmd
				break
			}
			window.Panes = append(window.Panes, pane)
		}
		config.Windows = append(config.Windows, window)
	}

	return config, nil
}
