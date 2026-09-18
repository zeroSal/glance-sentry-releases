package logger

import (
	"io"
	"os"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original }()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer error = %v", err)
	}

	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read output error = %v", err)
	}
	return string(out)
}

func TestFxErrorLoggerPrintf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
		want   string
	}{
		{
			name:   "error is printed in red",
			format: "[Fx] ERROR %s",
			args:   []any{"boom"},
			want:   "\033[31m[Fx] ERROR boom\033[0m\n",
		},
		{
			name:   "non error is dropped",
			format: "[Fx] PROVIDE %s",
			args:   []any{"*config.Env"},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				FxErrorLogger{}.Printf(tt.format, tt.args...)
			})

			if got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
