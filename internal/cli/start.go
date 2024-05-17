package cli

import (
	"fmt"

	"github.com/rafi/jig/pkg/client"
)

type StartCmd struct {
	Project   string            `help:"Project name to stop." arg:"" optional:""`
	Variables map[string]string `help:"Variable to interpolate in session config." arg:"" optional:""`
	Windows   []string          `help:"List of windows to start. If session exists, those windows will be attached to current session." short:"w" sep:","`
}

// Run executes the start command.
func (c *StartCmd) Run(jig client.Jig) error {
	configPath, err := FindProjectFile(c.Project, jig.Options.File)
	if err != nil {
		return err
	}
	config, err := client.LoadConfig(configPath, c.Variables)
	if err != nil {
		return err
	}

	if len(c.Windows) == 0 {
		fmt.Printf("Starting %q session…\n", shortenPath(configPath))
	} else {
		fmt.Printf("Creating %q new windows…\n", shortenPath(configPath))
	}
	return jig.Start(config, c.Windows)
}
