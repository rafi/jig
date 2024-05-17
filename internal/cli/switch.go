package cli

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/rafi/jig/pkg/client"
	"github.com/rafi/jig/pkg/fzf"
	"github.com/rafi/jig/pkg/tmux"
)

type SwitchCmd struct{
	Session string `arg:"" optional:"" help:"Optional session name to switch to."`
}

// Run executes the switch command.
func (c *SwitchCmd) Run(jig client.Jig) error {
	if c.Session != "" {
		return jig.SwitchOrAttach(c.Session)
	}

	sessions, err := jig.Tmux.ListSessions()
	if err != nil {
		return err
	}

	buffer := bytes.Buffer{}
	for _, session := range sessions {
		buffer.WriteString(formatSessionStatus(session, jig.Theme))
	}

	// Run fzf with the sub-command 'list' as preview.
	finder := fzf.New(jig.Theme.FzfArgs...)
	finder.WithPrompt("SESSION> ")

	selection, err := finder.Run(buffer)
	if err != nil {
		return err
	}
	if selection == "" {
		return nil
	}
	sessionID := strings.Split(selection, " ")[0]

	// Attach/switch to the session.
	if jig.Options.Detach {
		return nil
	}
	return jig.SwitchOrAttach(sessionID)
}

// formatSessionStatus returns a string representation of a tmux session.
func formatSessionStatus(session tmux.TmuxSession, theme client.Theme) string {
	marked := ""
	if session.Marked {
		marked = theme.IconMarked
	}
	attached := ""
	if session.Attached {
		attached = theme.IconAttached
	}
	alerts := ""
	if session.Alerts != "" {
		alerts = theme.IconAlert
	}
	status := fmt.Sprintf(
		"%s%s%s",
		theme.Attached.Copy().Width(2).Render(attached),
		theme.Activity.Copy().Width(2).Render(alerts),
		theme.Marked.Copy().Render(marked),
	)
	dates := fmt.Sprintf("(created @ %s, attached @ %s)",
		parseDate(session.Created),
		parseDate(session.LastAttached),
	)
	return fmt.Sprintf(
		"%s %s %s %s\n",
		theme.ID.Copy().Width(10).Render(session.Name),
		status,
		theme.Windows.Render(strconv.Itoa(session.Windows)+theme.IconTmux),
		theme.Date.Render(dates),
	)
}
