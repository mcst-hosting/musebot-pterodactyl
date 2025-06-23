package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var PythonNotFound = errors.New("python not found")

func Python(options BasicOptions, args ...string) (output string, err error) {

	if _, err := Which("python3", BasicOptions{
		Env:     options.Env,
		Sources: options.Sources,
		Cwd:     options.Cwd,
	}); err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", PythonNotFound
		} else {
			return "", err
		}
	}

	command := exec.Command("python3", args...)

	outputBytes, err := command.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("command error: %w", err)
		}
	}

	return strings.Trim(string(outputBytes), "\n"), nil

}
