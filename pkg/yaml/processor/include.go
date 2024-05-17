package processor

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/rafi/jig/pkg/shell"
)

// used for loading included files
type Fragment struct {
	content *yaml.Node
}

func (f *Fragment) UnmarshalYAML(value *yaml.Node) error {
	var err error
	// process includes in fragments
	f.content, err = resolveIncludes(value)
	return err
}

type IncludeProcessor struct {
	Out interface{}
}

func (i *IncludeProcessor) UnmarshalYAML(value *yaml.Node) error {
	resolved, err := resolveIncludes(value)
	if err != nil {
		return err
	}
	return resolved.Decode(i.Out)
}

func resolveIncludes(node *yaml.Node) (*yaml.Node, error) {
	if node.Tag == "!include" {
		if node.Kind != yaml.ScalarNode {
			return nil, errors.New("!include on a non-scalar node")
		}
		includePath := shell.ExpandPath(node.Value)
		file, err := os.ReadFile(includePath)
		if err != nil {
			return nil, err
		}
		var f Fragment
		err = yaml.Unmarshal(file, &f)
		f.content.Content = append(
			f.content.Content,
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: "config_path",
			},
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: node.Value,
				Style: yaml.DoubleQuotedStyle,
			},
		)
		return f.content, err
	}
	if node.Kind == yaml.SequenceNode || node.Kind == yaml.MappingNode {
		var err error
		for i := range node.Content {
			node.Content[i], err = resolveIncludes(node.Content[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return node, nil
}
