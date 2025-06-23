package linux

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

func NewCommand(options CommandOptions) (*LinuxCommand, error) {

	if len(options.Shell) == 0 {
		options.Shell = "/bin/bash"
	}

	if len(options.Cwd) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, ErrFetchingCwd
		}
		options.Cwd = cwd
	}

	return &LinuxCommand{
		Options:  options,
		handlers: make(map[int]interface{}),
	}, nil

}

func (cmd *LinuxCommand) AddHandler(handler interface{}) error {

	switch h := handler.(type) {
	case func(data EventOutputData) error:
		cmd.handlers[EventOutput] = h
		break
	case func(data EventExitData) error:
		cmd.handlers[EventExit] = h
		break
	default:
		return ErrInvalidHandler
	}

	return nil

}

func (cmd *LinuxCommand) Run() error {

	var sourceCommand strings.Builder
	for _, value := range cmd.Options.Sources {
		sourceCommand.WriteString(fmt.Sprintf("source %s && ", value))
	}

	var commandOptions strings.Builder
	commandOptions.WriteString(" ")
	for index, arg := range cmd.Options.Args {
		if len(cmd.Options.Args)-1 == index {
			commandOptions.WriteString(fmt.Sprintf("%s", arg))
		} else {
			commandOptions.WriteString(fmt.Sprintf("%s ", arg))
		}
	}

	command := exec.Command(cmd.Options.Shell, "-c", sourceCommand.String()+cmd.Options.Command+commandOptions.String())
	command.SysProcAttr = &unix.SysProcAttr{Setsid: true}
	command.Dir = cmd.Options.Cwd

	for key, value := range cmd.Options.Env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", key, value))
	}

	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, unix.SIGINT, unix.SIGTERM)

	var err error
	cmd.stdout, err = command.StdoutPipe()
	if err != nil {
		return err
	}

	cmd.stdin, err = command.StdinPipe()
	if err != nil {
		return err
	}

	cmd.stderr, err = command.StderrPipe()
	if err != nil {
		return err
	}

	if cmd.Options.PrintOutput {
		cmd.wg.Add(2)
		go func() {
			defer cmd.wg.Done()
			io.Copy(os.Stdout, cmd.stdout)
		}()
		go func() {
			defer cmd.wg.Done()
			io.Copy(os.Stderr, cmd.stderr)
		}()
	}

	if err := command.Start(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 127 {
				return ErrCommandNotFound
			} else if _, ok := cmd.Options.CustomErrors[int8(exitErr.ExitCode())]; ok {
				return cmd.Options.CustomErrors[int8(exitErr.ExitCode())]
			} else {
				return fmt.Errorf("%s: %w", ErrRunningCmd.Error(), err)
			}
		}
	}

	if len(cmd.handlers) != 0 {
		cmd.wg.Add(2)

		go func() {
			defer cmd.wg.Done()
			scanner := bufio.NewScanner(cmd.stderr)
			for scanner.Scan() {
				line := scanner.Text()
				if h, ok := cmd.handlers[EventOutput]; ok {
					if err := h.(func(data EventOutputData) error)(EventOutputData{
						Output:     line,
						CmdOptions: cmd.Options,
					}); err != nil {
						return
					}
				}
			}
		}()

		go func() {
			defer cmd.wg.Done()

			scanner := bufio.NewScanner(cmd.stdout)
			for scanner.Scan() {
				line := scanner.Text()
				if h, ok := cmd.handlers[EventOutput]; ok {
					if err := h.(func(data EventOutputData) error)(EventOutputData{
						Output:     line,
						CmdOptions: cmd.Options,
					}); err != nil {
						return
					}
				}
			}
		}()
	}

	cmd.wg.Add(1)

	go func() {
		defer cmd.wg.Done()

		select {
		case _, ok := <-signalChannel:
			if !ok {
				return
			}
			if err := unix.Kill(-command.Process.Pid, syscall.SIGINT); err != nil {
				return
			}
		}
	}()

	var exitInfo *EventExitData

	if _, ok := cmd.handlers[EventExit]; ok {
		exitInfo = &EventExitData{
			HasSucceeded: true,
			CmdOptions:   cmd.Options,
		}
	}

	if err := command.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.String() != "signal: interrupt" {
			close(signalChannel)
			cmd.wg.Wait()

			if exitErr.ExitCode() == 127 {
				return ErrCommandNotFound
			} else if _, ok := cmd.Options.CustomErrors[int8(exitErr.ExitCode())]; ok {
				if h, ok := cmd.handlers[EventExit]; ok {
					if exitInfo == nil {
						return fmt.Errorf("%s: %w", ErrRunningCmd.Error(), err)
					}
					exitInfo.HasSucceeded = false
					exitInfo.ExitCode = exitErr.ExitCode()
					var stdoutData bytes.Buffer
					if _, err := io.Copy(&stdoutData, cmd.stdout); err != nil {
						return err
					}
					exitInfo.Error = stdoutData.String()
					err := h.(func(data EventExitData) error)(*exitInfo)
					if err != nil {
						return fmt.Errorf("%s: %w", ErrRunningEvt.Error(), err)
					}
				}
				return cmd.Options.CustomErrors[int8(exitErr.ExitCode())]
			} else {
				if h, ok := cmd.handlers[EventExit]; ok {
					if exitInfo == nil {
						return fmt.Errorf("%s: %w", ErrRunningEvt.Error(), err)
					}
					exitInfo.HasSucceeded = false
					exitInfo.ExitCode = exitErr.ExitCode()
					var stdoutData bytes.Buffer
					if _, err := io.Copy(&stdoutData, cmd.stdout); err != nil {
						return err
					}
					exitInfo.Error = stdoutData.String()
					err := h.(func(data EventExitData) error)(*exitInfo)
					if err != nil {
						return fmt.Errorf("%s: %w", ErrRunningEvt.Error(), err)
					}
				}
				return fmt.Errorf("%s: %w", ErrRunningCmd.Error(), err)
			}
		}
	}

	if h, ok := cmd.handlers[EventExit]; ok {
		if exitInfo == nil {
			return nil
		}
		exitInfo.ExitCode = 0
		err := h.(func(data EventExitData) error)(*exitInfo)
		if err != nil {
			return fmt.Errorf("%s: %w", ErrRunningEvt.Error(), err)
		}
	}

	signal.Stop(signalChannel)
	close(signalChannel)
	cmd.wg.Wait()

	return nil

}
