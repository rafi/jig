package client_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rafi/jig/pkg/client"
	"github.com/rafi/jig/pkg/tmux"
)

func TestPrintCurrentSession(t *testing.T) {
	defaultShell := filepath.Base(os.Getenv("SHELL"))

	expectedConfig := client.Config{
		Session: "foobar",
		Path:    "/root",
		Windows: []client.Window{
			{
				Name:   "win1",
				Layout: "layout",
				Panes: []client.Pane{
					{
						Path: "/opt",
					},
					{
						Path: "/tmp",
						Cmd:  "nvim",
					},
				},
			},
		},
	}

	commander := &MockCommander{
		Commands: []string{},
		Outputs: []string{
			"foobar",
			strings.Join([]string{"id1", "win1", "layout", "/root"}, tmux.ColumnSep),
			strings.Join([]string{
				strings.Join([]string{"/opt", defaultShell}, tmux.ColumnSep),
				strings.Join([]string{"/tmp", "nvim"}, tmux.ColumnSep),
			}, "\n"),
		},
	}
	client := client.Jig{
		Tmux:      tmux.TmuxClient{Cmd: commander},
		Options:   client.Options{},
		InSession: false,
	}
	actualConfig, err := client.GenerateSessionConfig("test")
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, actualConfig)
}
