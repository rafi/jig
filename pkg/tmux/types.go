package tmux

import (
	"errors"
	"fmt"
	"time"
)

const (
	LayoutTiled          string = "tiled"
	LayoutMainHorizontal string = "main-horizontal"
	LayoutEvenHorizontal string = "even-horizontal"
	LayoutMainVertical   string = "main-vertical"
	LayoutEvenVertical   string = "even-vertical"
)

var (
	ErrInvalidSplitType = errors.New("invalid split type")
	ErrInvalidFormat    = errors.New("invalid shell output format")
)

type Target struct {
	Session string
	Window  string
	Pane    string
}

func (t Target) Get() string {
	target := fmt.Sprintf("%s:%s", t.Session, t.Window)
	if t.Pane == "" {
		return target
	}
	return target + "." + t.Pane
}

type TmuxSession struct {
	ID           string    `format:"session_id"`
	Name         string    `format:"session_name"`
	Path         string    `format:"session_path"`
	Attached     bool      `format:"session_attached"`
	Marked       bool      `format:"session_marked"`
	Windows      int       `format:"session_windows"`
	Stack        string    `format:"session_stack"`
	Alerts       string    `format:"session_alerts"`
	Created      time.Time `format:"session_created"`
	Activity     time.Time `format:"session_activity"`
	LastAttached time.Time `format:"session_last_attached"`
}

type TmuxWindow struct {
	ID     string `format:"window_id"`
	Name   string `format:"window_name"`
	Layout string `format:"window_layout"`
	Path   string `format:"pane_current_path"`
}

type TmuxPane struct {
	Path    string `format:"pane_current_path"`
	Command string `format:"pane_current_command"`
}
