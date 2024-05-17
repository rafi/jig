package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/rafi/jig/internal/cli"
	"github.com/rafi/jig/pkg/client"
	"github.com/rafi/jig/pkg/shell"
)

// newLogger creates a new logger instance.
func newLogger(path string) *log.Logger {
	logPath := filepath.Join(path, "jig.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(logFile, "", 0)
}

// main instantiates app and parse arguments.
func main() {
	os.Args = cli.ShimArgs(os.Args)
	cli, ctx := cli.NewApp()
	var logger *log.Logger
	if cli.Debug {
		logger = newLogger(filepath.Join(os.Getenv("HOME"), ".cache"))
	}
	cmd := shell.DefaultCommander{Logger: logger}

	jig, err := client.New(cli.Options, cmd)
	if err != nil {
		log.Fatal(err)
	}
	err = ctx.Run(jig)
	ctx.FatalIfErrorf(err)
}
