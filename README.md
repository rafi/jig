# Jig - tmux launcher

> Automates your [tmux](https://github.com/tmux/tmux) workflow.
> Inspired by [tmuxinator](https://github.com/tmuxinator/tmuxinator),
> [tmuxp](https://github.com/tmux-python/tmuxp), and
> [smug](https://github.com/ivaaaan/smug).

Define windows and panes in a single YAML file, and Jig will recreate it any
time you want.

Jig can also _**generate**_ a config file from your current tmux sessions, _**interpolate**_
custom variables, _**partially**_ start specific windows, and even include other
configuration files with the `!include` YAML directive.

## Features

- Recreate tmux sessions, windows, and panes from a single YAML file.
- Support YAML `!include <file>` directive to include other session files.
- Partially restore windows from configuration.
- Support variable interpolation in configurations.
- Generate current tmux session as YAML.
- Switch between sessions using fzf.

## Installation

- Download from the [releases page](https://github.com/rafi/jig/releases)
- Compile: `git clone git@github:/rafi/jig.git && cd jig && go install`

## Usage

```sh
jig <command> [project] [flags]
```

### Flags

```sh
  -h, --help           Show context-sensitive help.
      --debug          Print all commands to ~/.cache/jig.log
  -f, --file=STRING    Custom path to a config file
  -d, --detach         Detach tmux session. The same as -d flag in the tmux
  -i, --inside         Create all windows inside current session
```

### Configuration

Configuration files can stored in the `~/.config/jig` directory in `YAML`
format, e.g `~/.config/jig/your-project.yml`. You can use
`JIG_SESSION_CONFIG_PATH` to change the default base path if you wish.

You may also create a file named `.jig.yml` in the current working directory,
which will be used by default when no project name is provided.

You can use `!include <file>` directive to include other session files.
For example:

```yaml
sessions:
  - !include ~/code/a/.jig.yml
  - !include ~/code/b/.jig.yml
  - !include ~/code/c/.jig.yml
```

### User Variables

You can pass custom variables which will be interpolated with your configuration
files. First, use `${variable_name}` syntax in your config and then provide
key-value arguments:

```sh
$ cat ~/.config/jig/foo.yml
---
session: foo
windows:
  - cmd: echo ${year}

$ jig start foo year="$(date +%Y)"
```

This will create a window and run `echo 2024` within it.

### Examples

To create a new project, or edit an existing one with your `$EDITOR`:

```sh
jig new foo
jig edit foo
```

If you're already in a tmux session, you can generate it quickly:

```sh
jig print
jig print > .jig.yml
```

To start/stop a project and all windows, run:

```sh
jig start foo
jig stop foo
```

Also, jig commands have aliases:

```sh
jig foo # the same as "jig start foo"
jig st foo # the same as "jig stop foo"
jig p bar # the same as "jig print bar"
```

When you already have a running session, and you only want to create some
windows from the configuration file, use the `-w`, `--windows` flag or a
compound `<project:windows>` syntax and comma-separated window names:

```sh
# The following 3 commands are equivalent:
jig start project:window1,window2
jig start project -w window1,window2
jig start project -w window1 -w window2
```

You can use a custom path in the `-f` flag:

```sh
jig start -f ./project.yml
jig stop -f ./project.yml
jig start -f ./project.yml -w window1 -w window2
```

### Config Examples

#### Example 1

Showing all features:

```yaml
---
session: petstore
path: ~/code/petstore
env:
  FOO: BAR
before:
  # backend/docker-compose.yml is relative to session `path`
  - docker-compose -f backend/docker-compose.yml up -d
after:
  - docker stop $(docker ps -q)
sessions:
  - !include frontend/.jig.yml

windows:
  - name: code
    path: blog  # Relative path to session
    layout: main-vertical
    focus: true
    panes:
      - focus: true
        commands:
          - docker-compose start
      - type: horizontal
        commands:
          - sleep 4
          - docker-compose exec db psql
          - \dn; \dt public.*

  - name: run
    command:
      - git status -sb
      - git log --graph --all
        --pretty='%C(240)%h%C(reset) -%C(auto)%d%Creset %s %C(242)(%an %ar)'

  - name: infra
    path: ~/code/nlu
    layout: tiled
    manual: true  # Start this window only manually, using the -w argument.
    panes:
      - type: horizontal
        commands:
          - docker-compose up -d
```

#### Example 2

Short `cmd` example:

```yaml
---
session: blog
# When no path is provided, the config file's directory is used.
windows:
  - name: code
    layout: main-horizontal
    panes:
      - cmd: $EDITOR
      - cmd: make run-tests
  - name: ssh
    cmd: ssh myserver
```

## License

MIT License (c) 2024 Rafael Bodill
