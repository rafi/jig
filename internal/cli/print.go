package cli

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/rafi/jig/pkg/client"
)

const printIdent = 2

type PrintCmd struct {
	Session string `arg:"" optional:"" help:"Optional session name instead of current."`
}

// Run executes the print command.
func (c *PrintCmd) Run(jig client.Jig) error {
	config, err := jig.GenerateSessionConfig(c.Session)
	if err != nil {
		return err
	}

	raw, err := encodeYAML(config, printIdent)
	if err != nil {
		return err
	}

	fmt.Println(string(raw))
	return nil
}

// encodeYAML encodes the config to YAML with specified indentation count.
func encodeYAML(config client.Config, indent int) ([]byte, error) {
	var out bytes.Buffer
	e := yaml.NewEncoder(&out)
	e.SetIndent(indent)

	if err := e.Encode(&config); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
