package client

import "github.com/rafi/jig/pkg/tmux"

// Stop stops a tmux session and its nested sessions, if any.
func (j Jig) Stop(config Config, windows []string) error {
	for _, s := range config.Sessions {
		if err := j.stopSession(s, windows); err != nil {
			return err
		}
	}
	return j.stopSession(config, windows)
}

// stopSession stops a tmux session, and optionally run `after` commands.
func (j Jig) stopSession(session Config, windows []string) error {
	target := tmux.Target{Session: session.Session}

	if len(windows) == 0 {
		// Executes `after` commands.
		if len(session.After) > 0 {
			sessionPath, err := session.GetSessionPath()
			if err != nil {
				return err
			}
			if err := j.execShellCommands(session.After, sessionPath); err != nil {
				return err
			}
		}
		_, err := j.Tmux.StopSession(target)
		return err
	}

	// Kill specific windows
	for _, window := range windows {
		target.Window = window
		if err := j.Tmux.KillWindow(target); err != nil {
			return err
		}
	}
	return nil
}
