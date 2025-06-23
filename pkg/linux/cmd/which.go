package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrNotFound = errors.New("which: command not found")

type BasicOptions struct {
	Env     map[string]string
	Sources []string
	Cwd     string
}

func Which(cmd string, options BasicOptions) (dir string, err error) {

	var sourceCommand strings.Builder
	for _, value := range options.Sources {
		sourceCommand.WriteString(fmt.Sprintf("source %s && ", value))
	}

	command := exec.Command("which", cmd)

	if options.Cwd != "" {
		command.Dir = options.Cwd
	}

	for k, v := range options.Env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}

	outputBytes, err := command.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return "", ErrNotFound
			} else {
				return "", fmt.Errorf("command error: %w", err)
			}
		}
	}

	return strings.Trim(string(outputBytes), "\n"), nil

}
