package client

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/rafi/jig/pkg/shell"
	"github.com/rafi/jig/pkg/yaml/processor"
)

const (
	defaultCommandDelay         = 500
	envSessionVarName           = "JIG_SESSION"
	envSessionConfigPathVarName = "JIG_SESSION_CONFIG_PATH"
)

type Config struct {
	Session         string            `yaml:"session"`
	Env             map[string]string `yaml:"env,omitempty"`
	Path            string            `yaml:"path"`
	Before          []string          `yaml:"before,omitempty"`
	After           []string          `yaml:"after,omitempty"`
	Windows         []Window          `yaml:"windows"`
	CommandDelay    int               `yaml:"command_delay,omitempty"`
	SuppressHistory bool              `yaml:"suppress_history,omitempty"`
	Sessions        []Config          `yaml:"sessions,omitempty"`

	ConfigPath string `yaml:"config_path,omitempty"`
}

type Window struct {
	Name     string   `yaml:"name"`
	Before   []string `yaml:"before,omitempty"`
	Panes    []Pane   `yaml:"panes,omitempty"`
	Layout   string   `yaml:"layout"`
	Focus    bool     `yaml:"focus,omitempty"`
	Manual   bool     `yaml:"manual,omitempty"`
	Path     string   `yaml:"path,omitempty"`
	Commands []string `yaml:"commands,omitempty"`
	Cmd      string   `yaml:"cmd,omitempty"`
}

type Pane struct {
	Type     string   `yaml:"type,omitempty"`
	Path     string   `yaml:"path,omitempty"`
	Focus    bool     `yaml:"focus,omitempty"`
	Commands []string `yaml:"commands,omitempty"`
	Cmd      string   `yaml:"cmd,omitempty"`
}

func (c Config) GetSessionPath() (string, error) {
	// Resolve session start directory.
	// If session path is empty, use config path.
	// If session path is "." or "./", use current directory.
	switch c.Path {
	case "":
		return filepath.Dir(c.ConfigPath), nil
	case ".", "./":
		return os.Getwd()
	default:
		return shell.ExpandPath(c.Path), nil
	}
}

// FindConfig finds the config filename in the specified directory.
func FindConfig(dir, project string) (string, error) {
	configPath := filepath.Join(dir, project)
	for _, ext := range []string{".yml", ".yaml"} {
		if _, err := os.Stat(configPath + ext); err == nil {
			return configPath + ext, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return "", ErrConfigNotFound
}

// GetConfigPath returns the default base config path.
func GetConfigPath() (string, error) {
	configPath := ""
	if value, ok := os.LookupEnv(envSessionConfigPathVarName); ok {
		configPath = value
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configPath = filepath.Join(homeDir, ".config", "jig")
	}
	return configPath, nil
}

// ListConfigs returns a list of config files in the specified directory.
func ListConfigs(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}

	var result []string
	for _, file := range files {
		fileExt := path.Ext(file.Name())
		if fileExt != ".yml" && fileExt != ".yaml" {
			continue
		}
		result = append(result, file.Name())
	}
	return result, nil
}

// LoadConfig reads an entire config file, parses it with supplied variables,
// adds default environment variables and returns the final config.
func LoadConfig(path string, vars map[string]string) (Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	c, err := RenderConfig(string(f), vars)
	if err != nil {
		return c, err
	}

	// Resolve symlink path.
	if fi, _ := os.Lstat(path); fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return Config{}, err
		}
		if !filepath.IsAbs(realPath) {
			realPath = filepath.Clean(filepath.Join(filepath.Dir(path), realPath))
		}
		path = realPath
	}

	c.ConfigPath = path
	c.Env[envSessionVarName] = c.Session
	c.Env[envSessionConfigPathVarName] = path
	return c, err
}

// RenderConfig renders contents with supplied variables.
func RenderConfig(data string, vars map[string]string) (Config, error) {
	data = os.Expand(data, func(v string) string {
		if val, ok := vars[v]; ok {
			return val
		}

		if val, ok := os.LookupEnv(v); ok {
			return val
		}
		return v
	})

	c := Config{
		Env:          make(map[string]string),
		CommandDelay: defaultCommandDelay,
	}

	err := yaml.Unmarshal([]byte(data), &processor.IncludeProcessor{Out: &c})
	if err != nil {
		return Config{}, err
	}
	return c, nil
}

// GetEditor returns the editor to use.
func GetEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		for _, e := range []string{"nvim", "vim", "nano"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return "", ErrEditorNotFound
	}
	return editor, nil
}

// EditFile launches the default editor to edit a specified file.
func EditFile(path string) error {
	editor, err := GetEditor()
	if err != nil {
		return err
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
