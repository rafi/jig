package cli_test

import (
	"os"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"

	"github.com/rafi/jig/internal/cli"
	"github.com/rafi/jig/pkg/client"
)

func TestShimArgs(t *testing.T) {
	tests := []struct {
		args     []string
		expected []string
	}{
		{[]string{"jig"}, []string{"jig", "--help"}},
		{[]string{"jig", "foo"}, []string{"jig", "foo"}},
		{[]string{"jig", "start", "foo"}, []string{"jig", "start", "foo"}},
		{[]string{"jig", "stop", "foo:win"}, []string{"jig", "stop", "foo", "-w", "win"}},
		{[]string{"jig", "stop", "foo:win1,win2"}, []string{"jig", "stop", "foo", "-w", "win1,win2"}},
		{[]string{"jig", "foo:win"}, []string{"jig", "foo", "-w", "win"}},
		{[]string{"jig", "foo:win1", "-w", "win2"}, []string{"jig", "foo", "-w", "win2", "-w", "win1"}},
		{[]string{"jig", "foo:win1", "-w", "win2,win3"}, []string{"jig", "foo", "-w", "win2,win3", "-w", "win1"}},
	}

	t.Run("should shim help flag and split compound project:win1,win2", func(t *testing.T) {
		for _, v := range tests {
			args := cli.ShimArgs(v.args)
			assert.Equal(t, v.expected, args)
		}
	})
}

func TestCommandArgs(t *testing.T) {
	tests := []struct {
		argv []string
		wins []string
		opts client.Options
		err  error
		env  map[string]string
	}{
		{
			[]string{"start", "jig"},
			[]string{},
			client.Options{
				Detach: false,
				Debug:  false,
			},
			nil,
			nil,
		},
		{
			[]string{"start", "-d", "jig", "-w", "foo"},
			[]string{"foo"},
			client.Options{Detach: true},
			nil,
			nil,
		},
		{
			[]string{"start", "-d", "jig:foo,bar"},
			[]string{"foo", "bar"},
			client.Options{Detach: true},
			nil,
			nil,
		},
		{
			[]string{"start", "jig", "--debug", "--detach"},
			[]string{},
			client.Options{Debug: true, Detach: true},
			nil,
			nil,
		},
		{
			[]string{"start", "jig", "-d"},
			[]string{},
			client.Options{Detach: true},
			nil,
			nil,
		},
		{
			[]string{"start", "-df", "test.yml"},
			[]string{},
			client.Options{Detach: true, File: "test.yml"},
			nil,
			nil,
		},
		{
			[]string{"start", "-df", "test.yml", "-w", "win1", "-w", "win2"},
			[]string{"win1", "win2"},
			client.Options{Detach: true, File: "test.yml"},
			nil,
			nil,
		},
		{
			[]string{"start", "project", "a=b", "x=y"},
			[]string{},
			client.Options{},
			nil,
			nil,
		},
		{
			[]string{"start", "-f", "test.yml", "a=b", "x=y"},
			[]string{},
			client.Options{File: "test.yml"},
			nil,
			nil,
		},
		{
			[]string{"start", "-f", "test.yml", "-w", "win1", "-w", "win2", "a=b", "x=y"},
			[]string{"win1", "win2"},
			client.Options{File: "test.yml"},
			nil,
			nil,
		},
		{
			[]string{"test"},
			[]string{},
			client.Options{
				// Command:  "start",
				// Project:  "test",
				// Windows:  []string{},
				// Settings: map[string]string{},
			},
			nil,
			nil,
		},
		{
			[]string{"test", "-w", "win1", "-w", "win2", "a=b", "x=y"},
			[]string{"win1", "win2"},
			client.Options{
				// Command:  "start",
				// Project:  "test",
				// Settings: map[string]string{"a": "b", "x": "y"},
			},
			nil,
			nil,
		},
		{
			[]string{"test"},
			[]string{},
			client.Options{},
			nil,
			map[string]string{
				"JIG_SESSION_CONFIG_PATH": "test",
			},
		},
		// {
		// 	[]string{"start", "--test"},
		// 	[]string{},
		// 	client.Options{},
		// 	errors.New("unknown flag: --test"),
		// 	nil,
		// },
	}

	for _, v := range tests {
		for k, v := range v.env {
			os.Setenv(k, v)
		}
		os.Args = append([]string{"jig"}, v.argv...)
		opts := []kong.Option{
			kong.Exit(func(int) {}),
			kong.NoDefaultHelp(),
		}

		t.Run("should be created without any errors", func(t *testing.T) {
			app, ctx := cli.NewApp(opts...)
			assert.NotNil(t, ctx)
			assert.NotNil(t, app)
			assert.Equal(t, v.opts, app.Options)
		})
	}
}
