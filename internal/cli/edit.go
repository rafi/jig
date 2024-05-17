package cli

import (
	"errors"

	"github.com/rafi/jig/pkg/client"
)

type EditCmd struct {
	Project string `arg:"" optional:"" help:"Optional project name."`
}

// Run executes the edit command.
func (c *EditCmd) Run(jig client.Jig) error {
	configPath, err := FindProjectFile(c.Project, jig.Options.File)
	if err != nil && !errors.Is(err, ErrConfigNotFound{}) {
		return err
	}
	return client.EditFile(configPath)
}
