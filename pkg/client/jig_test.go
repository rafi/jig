package client_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rafi/jig/pkg/client"
	"github.com/rafi/jig/pkg/tmux"
)

var homeDir = os.Getenv("HOME")

type MockCommander struct {
	Commands []string
	Outputs  []string
}

func (c *MockCommander) Exec(cmd *exec.Cmd) (string, error) {
	c.Commands = append(c.Commands, strings.Join(cmd.Args, " "))

	output := ""
	if len(c.Outputs) > 1 {
		output, c.Outputs = c.Outputs[0], c.Outputs[1:]
	} else if len(c.Outputs) == 1 {
		output = c.Outputs[0]
	}

	return output, nil
}

func (c *MockCommander) ExecSilently(cmd *exec.Cmd) error {
	c.Commands = append(c.Commands, strings.Join(cmd.Args, " "))
	return nil
}

func TestDetectInTmuxSession(t *testing.T) {
	opts := client.Options{TmuxPath: "tmux"}
	testTable := []struct {
		env       map[string]string
		inSession bool
	}{
		{map[string]string{}, false},
		{map[string]string{"TMUX": ""}, true},
		{map[string]string{"TERM": "xterm"}, false},
		{map[string]string{"TERM": "xterm", "TMUX": ""}, true},
	}

	for _, v := range testTable {
		os.Clearenv()
		for key, value := range v.env {
			os.Setenv(key, value)
		}
		j, err := client.New(opts, &MockCommander{})
		assert.NoError(t, err)
		assert.Equal(t, j.InSession, v.inSession)
	}
	os.Clearenv()
}

func TestStartStopSession(t *testing.T) {
	testTable := map[string]struct {
		client           client.Jig
		config           client.Config
		windows          []string
		startCommands    []string
		stopCommands     []string
		commanderOutputs []string
	}{
		"test with 1 window": {
			client.Jig{Options: client.Options{}},
			client.Config{
				Session: "ses",
				Path:    "~/root",
				Before:  []string{"command1", "command2"},
				Windows: []client.Window{
					{
						Name:     "win1",
						Commands: []string{"command1"},
					},
				},
			},
			[]string{},
			[]string{
				"tmux has-session -t ses:",
				"/bin/sh -c command1",
				"/bin/sh -c command2",
				"tmux new-session -Pd -F #{session_id} -s ses -n win1 -c " + homeDir + "/root",
				"tmux send-keys -t ses:win1 -l command1",
				"tmux send-keys -t ses:win1 Enter",
				"tmux attach -d -t ses",
			},
			[]string{
				"tmux kill-session -t ses:",
			},
			[]string{"ses", "win1"},
		},
		"test with 1 window and Detach: true": {
			client.Jig{Options: client.Options{Detach: true}},
			client.Config{
				Session: "ses",
				Path:    "/root",
				Before:  []string{"command1", "command2"},
				Windows: []client.Window{
					{
						Name: "win1",
					},
				},
			},
			[]string{},
			[]string{
				"tmux has-session -t ses:",
				"/bin/sh -c command1",
				"/bin/sh -c command2",
				"tmux new-session -Pd -F #{session_id} -s ses -n win1 -c /root",
			},
			[]string{
				"tmux kill-session -t ses:",
			},
			[]string{"xyz"},
		},
		"test with multiple windows and panes": {
			client.Jig{Options: client.Options{}},
			client.Config{
				Session: "ses",
				Path:    "/tmp",
				Windows: []client.Window{
					{
						Name:   "win1",
						Manual: false,
						Layout: "main-horizontal",
						Panes: []client.Pane{
							{
								Type:     "horizontal",
								Commands: []string{"command1"},
							},
						},
					},
					{
						Name:   "win2",
						Manual: true,
						Layout: "tiled",
					},
				},
				After: []string{
					"stop1",
					"stop2 -d --foo=bar",
				},
			},
			[]string{},
			[]string{
				"tmux has-session -t ses:",
				"tmux new-session -Pd -F #{session_id} -s ses -n win1 -c /tmp",
				"tmux split-window -Pd -t ses:win1 -h -c /tmp -F #{pane_id}",
				"tmux send-keys -t ses:win1.1 -l command1",
				"tmux send-keys -t ses:win1.1 Enter",
				"tmux select-layout -t ses:win1 main-horizontal",
				"tmux attach -d -t ses",
			},
			[]string{
				"/bin/sh -c stop1",
				"/bin/sh -c stop2 -d --foo=bar",
				"tmux kill-session -t ses:",
			},
			[]string{"ses", "win1", "1"},
		},
		"test start windows from option's Windows parameter": {
			client.Jig{},
			client.Config{
				Session: "ses",
				Path:    "/tmp",
				Windows: []client.Window{
					{
						Name:   "win1",
						Manual: false,
					},
					{
						Name:   "win2",
						Manual: true,
					},
				},
			},
			[]string{"win2"},
			[]string{
				"tmux has-session -t ses:",
				"tmux new-session -Pd -F #{session_id} -s ses -n win1 -c /tmp",
				"tmux new-window -Pd -t ses: -n win2 -F #{window_id} -c /tmp",
				"tmux attach -d -t ses",
			},
			[]string{
				"tmux kill-window -t ses:win2",
			},
			[]string{"xyz"},
		},
		"test attach to the existing session": {
			client.Jig{},
			client.Config{
				Session: "ses",
				Path:    "/tmp",
				Windows: []client.Window{
					{Name: "win1"},
				},
			},
			[]string{},
			[]string{
				"tmux has-session -t ses:",
				"tmux attach -d -t ses",
			},
			[]string{
				"tmux kill-session -t ses:",
			},
			[]string{""},
		},
		"test start a new session from another tmux session": {
			client.Jig{InSession: true},
			client.Config{
				Session: "ses",
				Path:    "/tmp",
			},
			[]string{},
			[]string{
				"tmux has-session -t ses:",
				"tmux new-session -Pd -F #{session_id} -s ses -c /tmp",
				"tmux switch-client -t ses",
			},
			[]string{
				"tmux kill-session -t ses:",
			},
			[]string{"xyz"},
		},
		"test switch a client from another tmux session": {
			client.Jig{InSession: true},
			client.Config{
				Session: "ses",
				Path:    "/tmp",
				Windows: []client.Window{
					{Name: "win1"},
				},
			},
			[]string{},
			[]string{
				"tmux has-session -t ses:",
				"tmux switch-client -t ses",
			},
			[]string{
				"tmux kill-session -t ses:",
			},
			[]string{""},
		},
		"test create new windows in current session with same name": {
			client.Jig{Options: client.Options{Inside: true}, InSession: true},
			client.Config{
				Session: "ses",
				Path:    "/tmp",
				Windows: []client.Window{
					// FIX:
					// {Name: "win1"},
					{Name: "win1"},
				},
			},
			[]string{},
			[]string{
				"tmux display-message -p #S",
				"tmux has-session -t ses:",
				"tmux new-window -Pd -t ses: -n win1 -F #{window_id} -c /tmp",
			},
			[]string{
				"tmux kill-session -t ses:",
			},
			[]string{"ses", ""},
		},
		"test create new windows in current session with different name": {
			client.Jig{Options: client.Options{Inside: true}, InSession: true},
			client.Config{
				Session: "ses",
				Path:    "/tmp",
				Windows: []client.Window{
					{ Name:   "win1" },
				},
			},
			[]string{},
			[]string{
				"tmux display-message -p #S",
				"tmux has-session -t ses:",
				"tmux new-window -Pd -t ses: -n win1 -F #{window_id} -c /tmp",
			},
			[]string{
				"tmux kill-session -t ses:",
			},
			[]string{"ses", "win1"},
		},
	}

	for testDescription, params := range testTable {
		t.Run("start session: "+testDescription, func(t *testing.T) {
			commander := &MockCommander{[]string{}, params.commanderOutputs}
			params.client.Tmux = tmux.TmuxClient{Bin: "tmux", Cmd: commander}
			assert.NoError(t, params.client.Start(params.config, params.windows))
			assert.Equal(t, params.startCommands, commander.Commands)
		})

		t.Run("stop session: "+testDescription, func(t *testing.T) {
			commander := &MockCommander{[]string{}, params.commanderOutputs}
			params.client.Tmux = tmux.TmuxClient{Bin: "tmux", Cmd: commander}
			assert.NoError(t, params.client.Stop(params.config, params.windows))
			assert.Equal(t, params.stopCommands, commander.Commands)
		})
	}
}
