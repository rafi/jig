package cli

import (
	"fmt"

	"github.com/rafi/jig/internal/version"
	"github.com/rafi/jig/pkg/client"
)

type VersionCmd struct{}

// Run executes the version command.
func (c *VersionCmd) Run(jig client.Jig) error {
	fmt.Printf("%s %s\n", appName, version.GetVersion())
	return nil
}
