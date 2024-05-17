package shell_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/rafi/jig/pkg/shell"
)

func TestMain(m *testing.M) {
	switch os.Getenv("JIG_TEST_COMMANDER") {
	case "":
		os.Exit(m.Run())
	case "echo":
		fmt.Println(strings.Join(os.Args[1:], " "))
	case "exit":
		os.Exit(42)
	}
}

func TestExec(t *testing.T) {
	logger := log.New(bytes.NewBuffer([]byte{}), "", 0)
	commander := shell.DefaultCommander{logger}

	cmd := exec.Command(os.Args[0], "42")
	cmd.Env = append(os.Environ(), "JIG_TEST_COMMANDER=echo")

	output, err := commander.Exec(cmd)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if output != "42" {
		t.Errorf("expected 42, got %q", output)
	}
}

func TestExecError(t *testing.T) {
	logger := log.New(bytes.NewBuffer([]byte{}), "", 0)
	commander := shell.DefaultCommander{logger}

	cmd := exec.Command(os.Args[0], "42")
	cmd.Env = append(os.Environ(), "JIG_TEST_COMMANDER=exit")

	_, err := commander.Exec(cmd)
	if err == nil {
		t.Errorf("expected error")
	}

	got := cmd.ProcessState.ExitCode()
	if got != 42 {
		t.Errorf("expected %d, got %d", 42, got)
	}
}

func TestExecSilently(t *testing.T) {
	logger := log.New(bytes.NewBuffer([]byte{}), "", 0)
	commander := shell.DefaultCommander{logger}

	cmd := exec.Command(os.Args[0], "42")
	cmd.Env = append(os.Environ(), "JIG_TEST_COMMANDER=echo")

	err := commander.ExecSilently(cmd)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestExecSilentlyError(t *testing.T) {
	logger := log.New(bytes.NewBuffer([]byte{}), "", 0)
	commander := shell.DefaultCommander{logger}

	cmd := exec.Command(os.Args[0], "42")
	cmd.Env = append(os.Environ(), "JIG_TEST_COMMANDER=exit")

	err := commander.ExecSilently(cmd)
	if err == nil {
		t.Errorf("expected error")
	}

	got := cmd.ProcessState.ExitCode()
	if got != 42 {
		t.Errorf("expected %d, got %d", 42, got)
	}
}
