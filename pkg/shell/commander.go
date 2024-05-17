package shell

import (
	"log"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var _ Commander = DefaultCommander{}

type DefaultCommander struct {
	Logger *log.Logger
}

// Exec executes a command and returns its output.
func (c DefaultCommander) Exec(cmd *exec.Cmd) (string, error) {
	if c.Logger != nil {
		c.Logger.Println(strings.Join(cmd.Args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if c.Logger != nil {
			c.Logger.Println(err, string(output))
		}
		return "", &ShellError{strings.Join(cmd.Args, " "), err}
	}

	return strings.TrimSuffix(string(output), "\n"), nil
}

// ExecSilently executes a command without returning its output.
func (c DefaultCommander) ExecSilently(cmd *exec.Cmd) error {
	if c.Logger != nil {
		c.Logger.Println(strings.Join(cmd.Args, " "))
	}

	err := cmd.Run()
	if err != nil {
		if c.Logger != nil {
			c.Logger.Println(err)
		}
		return &ShellError{strings.Join(cmd.Args, " "), err}
	}
	return nil
}

// ExpandPath expands a relative or home-relative path to an absolute path.
func ExpandPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if strings.HasPrefix(path, "~/") {
		user, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(user.HomeDir, path[2:])
	}
	abspath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abspath
}
