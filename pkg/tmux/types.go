package tmux

import (
	"errors"
	"time"
)

const (
	LayoutTiled          string = "tiled"
	LayoutMainHorizontal string = "main-horizontal"
	LayoutEvenHorizontal string = "even-horizontal"
	LayoutMainVertical   string = "main-vertical"
	LayoutEvenVertical   string = "even-vertical"

	VSplit SplitType = "-v"
	HSplit SplitType = "-h"
)

type SplitType string

var (
	ErrInvalidSplitType = errors.New("invalid split type")
	ErrInvalidFormat    = errors.New("invalid shell output format")
)

type TmuxSession struct {
	ID           string
	Name         string
	Path         string
	Attached     bool
	Marked       bool
	Windows      int
	Stack        string
	Alerts       string
	Created      time.Time
	Activity     time.Time
	LastAttached time.Time
}

type TmuxWindow struct {
	ID     string
	Name   string
	Layout string
	Path   string
}

type TmuxPane struct {
	Path    string
	Command string
}
