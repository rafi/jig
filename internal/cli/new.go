package cli

import (
	"errors"

	"github.com/rafi/jig/pkg/client"
)

type NewCmd struct {
	Project string `arg:"" optional:"" help:"Optional project name."`
}

// Run executes the new command.
func (c *NewCmd) Run(jig client.Jig) error {
	configPath, err := FindProjectFile(c.Project, jig.Options.File)
	if err != nil && !errors.Is(err, ErrConfigNotFound{}) {
		return err
	}
	return client.EditFile(configPath)
}
