package logger

import (
	"fmt"
	"glance-sentry-releases/app/ansi"
	"io"
	"os"
	"sync"
)

var _ LoggerInterface = (*ConsoleLogger)(nil)

type ConsoleLogger struct {
	out   io.Writer
	mutex sync.Mutex
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{out: os.Stderr}
}

func (l *ConsoleLogger) GetIdentifier() string {
	return "console"
}

func (l *ConsoleLogger) log(color, prefix, msg string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, err := fmt.Fprintf(
		l.out,
		"%s%s %s%s\n",
		color,
		prefix,
		msg,
		ansi.Reset,
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to console logger: %v\n", err)
	}
}

func (l *ConsoleLogger) Debug(msg string) {
	l.log(ansi.White, "[•]", msg)
}

func (l *ConsoleLogger) Info(msg string) {
	l.log(ansi.Blue, "[i]", msg)
}

func (l *ConsoleLogger) Warn(msg string) {
	l.log(ansi.Yellow, "[!]", msg)
}

func (l *ConsoleLogger) Error(msg string) {
	l.log(ansi.Red, "[×]", msg)
}

func (l *ConsoleLogger) Success(msg string) {
	l.log(ansi.Green, "[✓]", msg)
}

func (l *ConsoleLogger) List(msgs []string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, msg := range msgs {
		_, err := fmt.Fprintf(l.out, " · %s\n", msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to console logger: %v\n", err)
		}
	}
}
