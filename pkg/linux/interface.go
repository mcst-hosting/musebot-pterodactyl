package linux

import (
	"errors"
	"io"
	"sync"
)

type LinuxCommand struct {
	Options  CommandOptions
	handlers map[int]interface{}
	wg       sync.WaitGroup
	stdout   io.ReadCloser
	stderr   io.ReadCloser
	stdin    io.WriteCloser
}

type CommandOptions struct {
	Env          map[string]string
	Sources      []string
	Command      string
	Args         []string
	CustomErrors map[int8]error
	Cwd          string
	Shell        string
	PrintOutput  bool
}

// Errors
var (
	ErrFetchingCwd          = errors.New("error fetching cwd")
	ErrRunningCmd           = errors.New("error running command")
	ErrCommandNotFound      = errors.New("error command not found")
	ErrCommandNotExecutable = errors.New("error command not executable")
	ErrInvalidHandler       = errors.New("invalid handler")
	ErrRunningEvt           = errors.New("error running event")
)
