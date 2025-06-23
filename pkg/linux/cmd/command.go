package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func Command(cmd string) (string, error) {

	command := exec.Command("command", "-v", cmd)

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
