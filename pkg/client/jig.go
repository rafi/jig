package client

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rafi/jig/pkg/shell"
	"github.com/rafi/jig/pkg/tmux"
)

const DefaultConfigFile = ".jig.yml"

type Options struct {
	Debug    bool   `help:"Print all commands to ~/.cache/jig.log"`
	File     string `help:"Custom path to a config file." short:"f"`
	Detach   bool   `help:"Do not attach to the session." short:"d"`
	Inside   bool   `help:"Create windows inside current session." short:"i"`
	TmuxPath string
}

var (
	ErrConfigNotFound   = errors.New("project file not found")
	ErrEditorNotFound   = errors.New("editor not found")
	ErrNoWindowsFound   = errors.New("no windows found")
	ErrNoSessionName    = errors.New("you must specify a session name")
	ErrNotInsideSession = errors.New("cannot use -i flag outside of a tmux session")
)

type Jig struct {
	Tmux      tmux.TmuxClient
	Theme     Theme
	Options   Options
	InSession bool
}

// New creates a new Jig client.
func New(opts Options, commander shell.Commander) (Jig, error) {
	if opts.TmuxPath == "" {
		var err error
		opts.TmuxPath, err = exec.LookPath("tmux")
		if err != nil {
			return Jig{}, err
		}
	}
	tmux := tmux.TmuxClient{
		Bin: filepath.Clean(opts.TmuxPath),
		Cmd: commander,
	}
	_, inTmuxSession := os.LookupEnv("TMUX")

	return Jig{
		Tmux:      tmux,
		Options:   opts,
		Theme:     NewThemeDefault(),
		InSession: inTmuxSession,
	}, nil
}

// SwitchOrAttach switches to a tmux session or attaches to it if it exists.
func (j Jig) SwitchOrAttach(session string) error {
	if j.InSession {
		return j.Tmux.SwitchClient(session)
	}
	if !j.InSession {
		return j.Tmux.Attach(session, os.Stdin, os.Stdout, os.Stderr)
	}
	return nil
}

// execShellCommands executes a list of shell commands in a given directory.
func (j Jig) execShellCommands(commands []string, path string) error {
	path = shell.ExpandPath(path)
	for _, c := range commands {
		cmd := exec.Command("/bin/sh", "-c", c)
		cmd.Dir = path

		_, err := j.Tmux.Cmd.Exec(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sets a map of environment variables inside a tmux session.
func (j Jig) setEnvVariables(session string, env map[string]string) error {
	for key, value := range env {
		if _, err := j.Tmux.SetEnv(session, key, value); err != nil {
			return err
		}
	}
	return nil
}
