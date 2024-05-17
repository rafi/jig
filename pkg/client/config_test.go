package client_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/rafi/jig/pkg/client"
)

func TestRenderConfig(t *testing.T) {
	yaml := `
session: ${session}
command_delay: 200
windows:
  - layout: tiled
    commands:
      - echo 1
    panes:
      - commands:
        - echo 2
        - echo ${HOME}
        type: horizontal`

	config, err := client.RenderConfig(yaml, map[string]string{
		"session": "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := client.Config{
		Session:      "test",
		CommandDelay: 200,
		Env:          make(map[string]string),
		Windows: []client.Window{
			{
				Layout:   "tiled",
				Commands: []string{"echo 1"},
				Panes: []client.Pane{
					{
						Type:     "horizontal",
						Commands: []string{"echo 2", "echo " + os.Getenv("HOME")},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(expected, config) {
		t.Fatalf("expected %v, got %v", expected, config)
	}
}
