package cli

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/rafi/jig/pkg/client"
)

type ErrConfigNotFound struct{ Project, Path string }

func (e ErrConfigNotFound) Error() string {
	return fmt.Sprintf("config not found for project %s at %q", e.Project, e.Path)
}

// FindProjectFile parses the cli arguments and returns a runtime configuration.
func FindProjectFile(name, file string) (string, error) {
	var err error
	configPath := ""
	if file != "" {
		// Use the exact path from user.
		configPath = file
	} else if name == "" {
		configPath, err = getDefaultConfig()
	} else {
		configPath, err = getConfigPath(name)
	}
	return configPath, err
}

// If project name is not set, try to look for config file in current directory.
func getDefaultConfig() (string, error) {
	configPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	configPath = filepath.Join(configPath, client.DefaultConfigFile)
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrConfigNotFound{
				Project: filepath.Base(configPath),
				Path:    configPath,
			}
		}
		return "", err
	}
	return configPath, nil
}

// Look for a project config file in the global config path.
func getConfigPath(name string) (string, error) {
	configPath, err := client.GetConfigPath()
	if err != nil {
		return "", err
	}

	configPath, err = client.FindConfig(configPath, name)
	if err != nil {
		configPath = filepath.Join(configPath, name+".yml")
		return configPath, err
	}
	return configPath, nil
}

// shortenPath returns a path with user's home replaced to ~/
func shortenPath(path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Clean(path)
	}
	user, err := user.Current()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, user.HomeDir) {
		return filepath.Join("~/", path[len(user.HomeDir):])
	}
	return path
}

// parseDate returns a short formatted date string.
func parseDate(date time.Time) string {
	now := time.Now()
	dateFmt := "Jan-_2 15:04"
	if now.Month() == date.UTC().Month() {
		dateFmt = "_2 15:04"
	}
	if now.Day() == date.UTC().Day() {
		dateFmt = "15:04"
	}
	return date.Format(dateFmt)
}
