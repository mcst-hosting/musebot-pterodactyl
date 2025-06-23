package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var NodeNotFound = errors.New("nodejs not found")

func Node(options BasicOptions, args ...string) (output string, err error) {

	if _, err := Which("node", BasicOptions{
		Env:     options.Env,
		Sources: options.Sources,
		Cwd:     options.Cwd,
	}); err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", NodeNotFound
		} else {
			return "", err
		}
	}

	command := exec.Command("node", args...)

	outputBytes, err := command.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("command error: %w", err)
		}
	}

	return strings.Trim(string(outputBytes), "\n"), nil

}
