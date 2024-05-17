package client

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	ID       lipgloss.Style
	Date     lipgloss.Style
	Attached lipgloss.Style
	Marked   lipgloss.Style
	Activity lipgloss.Style
	Windows  lipgloss.Style

	IconMarked   string
	IconAttached string
	IconTmux     string
	IconAlert    string

	FzfArgs []string
}

// NewThemeDefault returns the default theme.
func NewThemeDefault() Theme {
	return Theme{
		ID:       lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")),
		Date:     lipgloss.NewStyle().Foreground(lipgloss.Color("#5A5A5A")),
		Attached: lipgloss.NewStyle().Foreground(lipgloss.Color("#84af00")),
		Marked:   lipgloss.NewStyle().Foreground(lipgloss.Color("#84afca")),
		Activity: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff8700")),
		Windows:  lipgloss.NewStyle().Foreground(lipgloss.Color("#008cc2")),

		IconMarked:   " ",
		IconAttached: "󰓾 ",
		IconTmux:     " ",
		IconAlert:    "󰆾 ",

		FzfArgs: []string{
			"--nth=1",
			"--height=~100%",
			"--separator= ",
			"--margin=1,5%",
			"--border",
			"--info=inline-right",
		},
	}
}
