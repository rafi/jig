package cli

import (
	"fmt"

	"github.com/rafi/jig/pkg/client"
)

type StopCmd struct {
	Project   string            `help:"Project name to start." arg:"" optional:""`
	Variables map[string]string `help:"Variable to interpolate in session config." arg:"" optional:""`
	Windows   []string          `help:"List of windows to stop." short:"w" sep:","`
}

// Run executes the stop command.
func (c *StopCmd) Run(jig client.Jig) error {
	configPath, err := FindProjectFile(c.Project, jig.Options.File)
	if err != nil {
		return err
	}
	config, err := client.LoadConfig(configPath, c.Variables)
	if err != nil {
		return err
	}

	if len(c.Windows) == 0 {
		fmt.Printf("Terminating %q session…\n", shortenPath(configPath))
	} else {
		fmt.Printf("Killing %q windows…\n", shortenPath(configPath))
	}
	return jig.Stop(config, c.Windows)
}
