package logger

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestConsoleLoggerFormatsColoredError(t *testing.T) {
	var output bytes.Buffer
	log := &consoleLogger{output: &output, color: true}

	log.log(LevelError, "database", "query failed", F("error", "line one\nline two"))

	line := output.String()
	for _, expected := range []string{
		"\x1b[31m\x1b[1mERROR",
		"DATABASE",
		"query failed",
		`error="line one\nline two"`,
	} {
		if !strings.Contains(line, expected) {
			t.Fatalf("expected log to contain %q, got %q", expected, line)
		}
	}
}

func TestDurationKeepsFastRequestsReadable(t *testing.T) {
	if got := Duration(240 * time.Microsecond); got != "<1ms" {
		t.Fatalf("expected sub-millisecond duration, got %q", got)
	}
	if got := Duration(1450 * time.Microsecond); got != "1ms" {
		t.Fatalf("expected rounded duration, got %q", got)
	}
}

func TestConsoleLoggerCanDisableColors(t *testing.T) {
	var output bytes.Buffer
	log := &consoleLogger{output: &output, color: false}

	log.log(LevelSuccess, "sync", "completed", F("saved", 10))

	if strings.Contains(output.String(), "\x1b[") {
		t.Fatalf("expected plain log without ANSI codes, got %q", output.String())
	}
	if !strings.Contains(output.String(), "saved=10") {
		t.Fatalf("expected structured field, got %q", output.String())
	}
}
