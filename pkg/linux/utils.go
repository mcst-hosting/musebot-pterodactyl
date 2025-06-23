package linux

import (
	"egtyl.xyz/omnibill/linux/cmd"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io/fs"
	"os"
	"os/exec"
)

func (c *LinuxCommand) isCommandExecutable(command string) (bool, error) {

	whichOut, err := cmd.Which(command, cmd.BasicOptions{
		Env:     c.Options.Env,
		Sources: c.Options.Sources,
		Cwd:     c.Options.Cwd,
	})
	if err != nil {
		if errors.Is(err, cmd.ErrNotFound) {
			if _, err := os.Stat(command); errors.Is(err, fs.ErrNotExist) {
				return false, err
			}
		} else {
			return false, err
		}
	}

	if len(whichOut) == 0 {
		return false, nil
	}

	if err := unix.Access(whichOut, unix.X_OK); err != nil {
		if err == unix.EACCES {
			return false, nil
		} else {
			fmt.Println(err)
			return false, err
		}
	}

	return true, nil

}

func (c *LinuxCommand) doesCommandExist(command string) (bool, error) {

	shellCmd := exec.Command(c.Options.Shell, "-c", fmt.Sprintf("command -v %s", command))

	if err := shellCmd.Start(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return false, nil
			} else {
				return false, ErrRunningCmd
			}
		}
	}

	if err := shellCmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return false, nil
			} else {
				return false, ErrRunningCmd
			}
		}
	}

	return true, nil

}
