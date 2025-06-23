package logger

import (
	"fmt"
	"strings"
)

type Logger struct {
	namePrefix string

	errorFormat   string
	warningFormat string
	infoFormat    string
}

var (
	namePrefixFormat = "\033[1m\033[38;5;221m[%s]\033[0m"
	formatError      = "\033[38;5;160m[ERROR]\033[0m"
	formatWarning    = "\033[38;5;214m[WARNING]\033[0m"
	formatInfo       = "\033[38;5;159m[INFO]\033[0m"
)

type Options struct {
	Prefix string

	ErrorFormat   string
	WarningFormat string
	InfoFormat    string
}

func New(options Options) *Logger {
	return &Logger{
		namePrefix: options.Prefix,

		errorFormat:   options.ErrorFormat,
		warningFormat: options.WarningFormat,
		infoFormat:    options.InfoFormat,
	}
}

func (logger *Logger) SetPrefix(prefix string) {
	logger.namePrefix = prefix
}

func (logger *Logger) SetErrorFormat(format string) {
	logger.errorFormat = format
}

func (logger *Logger) SetWarningFormat(format string) {
	logger.warningFormat = format
}

func (logger *Logger) SetInfoFormat(format string) {
	logger.infoFormat = format
}

func (logger *Logger) Info(message string) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(namePrefixFormat, logger.namePrefix) + " | ")
	builder.WriteString(formatInfo + " ")
	builder.WriteString(logger.infoFormat + message + "\033[0m")
	builder.WriteString("\n")

	fmt.Print(builder.String())
}

func (logger *Logger) Warn(message string) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(namePrefixFormat, logger.namePrefix) + " | ")
	builder.WriteString(formatWarning + " ")
	builder.WriteString(logger.warningFormat + message + "\033[0m")
	builder.WriteString("\n")

	fmt.Print(builder.String())
}

func (logger *Logger) Error(message string) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(namePrefixFormat, logger.namePrefix) + " | ")
	builder.WriteString(formatError + " ")
	builder.WriteString(logger.errorFormat + message + "\033[0m")
	builder.WriteString("\n")

	fmt.Print(builder.String())
}
