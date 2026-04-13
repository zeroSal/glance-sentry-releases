package logger

import (
	"fmt"
	"strings"
)

type FxErrorLogger struct{}

func (l FxErrorLogger) Printf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if idx := strings.Index(msg, "[Fx] ERROR"); idx != -1 {
		fmt.Printf("\033[31m%s\033[0m\n", msg)
	}
}
