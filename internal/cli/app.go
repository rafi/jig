package cli

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/rafi/jig/internal/version"
	"github.com/rafi/jig/pkg/client"
)

const (
	appName  = "jig"
	appDesc  = "tmux launcher"
	examples = `
$ jig list
$ jig edit foo
$ jig new foo
$ jig print > ~/.config/jig/foo.yml
$ jig foo
$ jig start foo
$ jig start foo -d
$ jig start foo:win1
$ jig start foo -w win1
$ jig start foo:win1,win2
$ jig stop foo
`
)

var _ = kong.Must(&CLI{})

type CLI struct {
	client.Options

	Start   StartCmd   `cmd:"" help:"Start a tmux session." aliases:"star,sta" default:"withargs"`
	Stop    StopCmd    `cmd:"" help:"Stop a tmux session." aliases:"sto"`
	Print   PrintCmd   `cmd:"" help:"Print the current tmux session's configuration." aliases:"pr,p"`
	List    ListCmd    `cmd:"" help:"List all projects, or project's windows." aliases:"l,ls"`
	Edit    EditCmd    `cmd:"" help:"Edit the a tmux session configuration." aliases:"ed,e"`
	New     NewCmd     `cmd:"" help:"Create a new tmux session." aliases:"ne,n"`
	Switch  SwitchCmd  `cmd:"" help:"Switch to existing tmux session." aliases:"swi,sw"`
	Version VersionCmd `cmd:"" help:"Display version information." aliases:"ver,v"`
}

// NewApp creates a new CLI application.
func NewApp(opts ...kong.Option) (CLI, *kong.Context) {
	cli := CLI{}
	ver := version.GetVersion()

	defaults := []kong.Option{
		kong.Name(appName),
		kong.Description(fmt.Sprintf("%s %s - %s", appName, ver, appDesc)),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Tree: true}),
		kong.Help(helpPrinter),
		kong.Vars{"version": ver},
	}

	// Combine defaults and user options and run the parser.
	if opts == nil {
		opts = []kong.Option{}
	}
	opts = append(defaults, opts...)
	ctx := kong.Parse(&cli, opts...)
	return cli, ctx
}

// helpPrinter prints the help message with examples.
func helpPrinter(options kong.HelpOptions, ctx *kong.Context) error {
	if err := kong.DefaultHelpPrinter(options, ctx); err != nil {
		return err
	}
	for _, line := range strings.Split(examples, "\n") {
		fmt.Println(strings.Repeat(" ", 2) + line)
	}
	return nil
}

// ShimArgs handle special cases when running the program:
// - If running without any arguments, default to the --help flag.
// - If running with a compound `project:windows` argument, split it.
func ShimArgs(args []string) []string {
	if len(args) < 2 {
		// If running without any extra arguments, default to the --help flag.
		return append(args, "--help")
	}
	parsed := []string{args[0]}
	windows := []string{}
	for _, arg := range args[1:] {
		if strings.Contains(arg, ":") {
			// Split compound `project:win1,win2` argument.
			pair := strings.Split(arg, ":")
			arg = pair[0]
			windows = append(windows, strings.Split(pair[1], ",")...)
		}
		parsed = append(parsed, arg)
	}
	if len(windows) > 0 {
		parsed = append(parsed, "-w", strings.Join(windows, ","))
	}
	return parsed
}
