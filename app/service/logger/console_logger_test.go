package logger

import (
	"bytes"
	"testing"

	"glance-sentry-releases/app/ansi"
)

func TestConsoleLoggerGetIdentifier(t *testing.T) {
	if got := NewConsoleLogger().GetIdentifier(); got != "console" {
		t.Errorf("GetIdentifier() = %q, want %q", got, "console")
	}
}

func TestConsoleLoggerLevels(t *testing.T) {
	tests := []struct {
		name string
		log  func(*ConsoleLogger, string)
		want string
	}{
		{"debug", (*ConsoleLogger).Debug, ansi.White + "[•] hello" + ansi.Reset + "\n"},
		{"info", (*ConsoleLogger).Info, ansi.Blue + "[i] hello" + ansi.Reset + "\n"},
		{"warn", (*ConsoleLogger).Warn, ansi.Yellow + "[!] hello" + ansi.Reset + "\n"},
		{"error", (*ConsoleLogger).Error, ansi.Red + "[×] hello" + ansi.Reset + "\n"},
		{"success", (*ConsoleLogger).Success, ansi.Green + "[✓] hello" + ansi.Reset + "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.log(&ConsoleLogger{out: &buf}, "hello")

			if got := buf.String(); got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConsoleLoggerList(t *testing.T) {
	var buf bytes.Buffer
	logger := &ConsoleLogger{out: &buf}

	logger.List([]string{"first", "second"})

	want := " · first\n · second\n"
	if got := buf.String(); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestConsoleLoggerListEmpty(t *testing.T) {
	var buf bytes.Buffer
	logger := &ConsoleLogger{out: &buf}

	logger.List(nil)

	if got := buf.String(); got != "" {
		t.Errorf("output = %q, want empty", got)
	}
}

func TestConsoleLoggerConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	logger := &ConsoleLogger{out: &buf}
	done := make(chan struct{})

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			logger.Info("hello")
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	want := len(ansi.Blue+"[i] hello"+ansi.Reset+"\n") * 10
	if got := buf.Len(); got != want {
		t.Errorf("written bytes = %d, want %d", got, want)
	}
}
