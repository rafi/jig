package cli

import (
	"fmt"
	"strings"

	"github.com/rafi/jig/internal/version"
	"github.com/rafi/jig/pkg/client"
)

type VersionCmd struct{}

// Run executes the version command.
func (c *VersionCmd) Run(jig client.Jig) error {
	info := version.Get()
	sha := info.GitCommit[0:8]
	v := info.Version
	if !strings.HasSuffix(info.Version, sha) {
		v += " " + sha
	}

	fmt.Printf("%s %s (%s)\n", appName, v, info.Date)
	return nil
}
