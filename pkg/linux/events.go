package linux

const (
	EventOutput = iota
	EventExit
)

type EventOutputData struct {
	Output     string
	CmdOptions CommandOptions
}

type EventExitData struct {
	HasSucceeded bool
	ExitCode     int
	CmdOptions   CommandOptions
	Error        string
}
