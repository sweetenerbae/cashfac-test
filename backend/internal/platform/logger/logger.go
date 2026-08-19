package logger

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Level string

const (
	LevelInfo    Level = "INFO"
	LevelSuccess Level = "SUCCESS"
	LevelWarn    Level = "WARN"
	LevelError   Level = "ERROR"
)

const (
	colorReset   = "\x1b[0m"
	colorBold    = "\x1b[1m"
	colorDim     = "\x1b[2m"
	colorRed     = "\x1b[31m"
	colorGreen   = "\x1b[32m"
	colorYellow  = "\x1b[33m"
	colorBlue    = "\x1b[34m"
	colorMagenta = "\x1b[35m"
	colorCyan    = "\x1b[36m"
)

type Field struct {
	Key   string
	Value any
}

var defaultLogger = &consoleLogger{
	output: os.Stdout,
	color:  colorEnabled(),
}

type consoleLogger struct {
	mu     sync.Mutex
	output io.Writer
	color  bool
}

func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Duration(value time.Duration) string {
	if value < time.Millisecond {
		return "<1ms"
	}
	return value.Round(time.Millisecond).String()
}

func Info(scope, message string, fields ...Field) {
	defaultLogger.log(LevelInfo, scope, message, fields...)
}

func Success(scope, message string, fields ...Field) {
	defaultLogger.log(LevelSuccess, scope, message, fields...)
}

func Warn(scope, message string, fields ...Field) {
	defaultLogger.log(LevelWarn, scope, message, fields...)
}

func Error(scope, message string, fields ...Field) {
	defaultLogger.log(LevelError, scope, message, fields...)
}

func Banner(title string, fields ...Field) {
	defaultLogger.banner(title, fields...)
}

func SetOutput(output io.Writer) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.output = output
}

func SetColor(enabled bool) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.color = enabled
}

func (l *consoleLogger) log(level Level, scope, message string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := paint(l.color, colorDim, time.Now().Format("15:04:05.000"))
	levelText := paint(l.color, levelColor(level), fmt.Sprintf("%-7s", level))
	scopeText := paint(l.color, colorForScope(scope), fmt.Sprintf("%-8s", strings.ToUpper(scope)))
	messageText := message
	if level == LevelError {
		messageText = paint(l.color, colorRed+colorBold, message)
	}

	line := fmt.Sprintf("%s  %s  %s  %s", timestamp, levelText, scopeText, messageText)
	if formatted := formatFields(fields); formatted != "" {
		line += "  " + paint(l.color, colorDim, formatted)
	}

	fmt.Fprintln(l.output, line)
}

func (l *consoleLogger) banner(title string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	const width = 62
	border := strings.Repeat("─", width)
	fmt.Fprintln(l.output, paint(l.color, colorCyan, "╭"+border+"╮"))
	writeBannerLine(l.output, l.color, width, strings.ToUpper(title), true)
	fmt.Fprintln(l.output, paint(l.color, colorCyan, "├"+border+"┤"))
	for _, field := range fields {
		label := fmt.Sprintf("%-10s", strings.ToUpper(field.Key))
		writeBannerField(l.output, l.color, width, label, bannerValue(field.Value))
	}
	fmt.Fprintln(l.output, paint(l.color, colorCyan, "╰"+border+"╯"))
}

func writeBannerLine(output io.Writer, color bool, width int, value string, bold bool) {
	content := "  " + value
	padding := max(width-len([]rune(content)), 0)
	textColor := colorCyan
	if bold {
		textColor = colorBold + colorCyan
	}
	fmt.Fprintf(output, "%s%s%s%s\n",
		paint(color, colorCyan, "│"),
		paint(color, textColor, content),
		strings.Repeat(" ", padding),
		paint(color, colorCyan, "│"),
	)
}

func writeBannerField(output io.Writer, color bool, width int, label, value string) {
	contentLength := 2 + len([]rune(label)) + 1 + len([]rune(value))
	padding := max(width-contentLength, 0)
	fmt.Fprintf(output, "%s  %s %s%s%s\n",
		paint(color, colorCyan, "│"),
		paint(color, colorDim, label),
		value,
		strings.Repeat(" ", padding),
		paint(color, colorCyan, "│"),
	)
}

func formatFields(fields []Field) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(field.Key) == "" {
			continue
		}
		parts = append(parts, field.Key+"="+sanitize(field.Value))
	}
	return strings.Join(parts, "  ")
}

func sanitize(value any) string {
	formatted := fmt.Sprint(value)
	if strings.ContainsAny(formatted, " \n\r\t") {
		return strconv.Quote(formatted)
	}
	return formatted
}

func bannerValue(value any) string {
	return strings.NewReplacer("\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(fmt.Sprint(value))
}

func paint(enabled bool, color, value string) string {
	if !enabled {
		return value
	}
	return color + value + colorReset
}

func levelColor(level Level) string {
	switch level {
	case LevelSuccess:
		return colorGreen + colorBold
	case LevelWarn:
		return colorYellow + colorBold
	case LevelError:
		return colorRed + colorBold
	default:
		return colorBlue + colorBold
	}
}

func colorForScope(scope string) string {
	switch strings.ToUpper(scope) {
	case "HTTP":
		return colorCyan
	case "LLM":
		return colorMagenta
	case "SYNC", "SOURCE":
		return colorGreen
	default:
		return colorBlue
	}
}

func colorEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_COLOR"))) {
	case "always", "true", "1":
		return true
	case "never", "false", "0":
		return false
	}

	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return false
	}

	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
