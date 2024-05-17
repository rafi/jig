package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rafi/jig/pkg/client"
	"github.com/xlab/treeprint"
)

type ListCmd struct {
	Project string `help:"Optional project name to list windows." arg:"" optional:""`
}

// Run executes the list command.
func (c *ListCmd) Run(jig client.Jig) error {
	if c.Project != "" {
		// List windows and panes of a single project, as a tree.
		configPath, err := FindProjectFile(c.Project, jig.Options.File)
		if err != nil {
			return err
		}
		config, err := client.LoadConfig(configPath, map[string]string{})
		if err != nil {
			return err
		}
		tree := displayConfigTree(config)
		tree.SetValue(config.Session)
		fmt.Print(tree.String())
		return nil
	}

	// List all projects.
	configPath, err := client.GetConfigPath()
	if err != nil {
		return err
	}
	configs, err := client.ListConfigs(configPath)
	if err != nil {
		return err
	}
	for _, config := range configs {
		fileExt := filepath.Ext(config)
		fmt.Println(strings.TrimSuffix(config, fileExt))
	}
	return nil
}

// makeTreeProject recursively builds a tree of a single project.
func displayConfigTree(project client.Config) treeprint.Tree {
	tree := treeprint.New()
	for _, session := range project.Sessions {
		branch := displayConfigTree(session)
		branch.SetValue(session.Session)
		tree.AddBranch(branch)
	}
	for _, win := range project.Windows {
		winBranch := treeprint.New()
		winBranch.SetValue(makeTreeWindowEntry(win))
		for _, pane := range win.Panes {
			winBranch.AddNode(makeTreePaneEntry(pane))
		}
		tree.AddNode(winBranch)
	}
	return tree
}

// makeTreeWindowEntry builds a tree entry for a window.
func makeTreeWindowEntry(win client.Window) string {
	if len(win.Cmd) > 0 {
		win.Commands = append(win.Commands, win.Cmd)
	}
	title := ""
	if len(win.Name) > 0 {
		title += fmt.Sprintf("[%s]", win.Name)
	}
	if len(win.Commands) > 0 {
		if title != "" {
			title += "\n"
		}
		title += strings.Join(win.Commands, "\n")
	}
	if len(win.Path) > 0 {
		title += title + " " + win.Path
	}
	return title
}

// makeTreePaneEntry builds a tree entry for a pane.
func makeTreePaneEntry(pane client.Pane) string {
	if len(pane.Cmd) > 0 {
		pane.Commands = append(pane.Commands, pane.Cmd)
	}
	title := ""
	if len(pane.Commands) > 0 {
		title += strings.Join(pane.Commands, "\n")
	}
	if len(pane.Path) > 0 {
		title += title + " " + pane.Path
	}
	return title
}
