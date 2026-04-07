package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

var Log *slog.Logger

// Init - init global logger
func Init(level slog.Level) {
	handler := &PrettyHandler{
		level: level,
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// --- PrettyHandler ---

type PrettyHandler struct {
	level slog.Level
}

func (h *PrettyHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	t := r.Time.Format("2006-01-02 15:04:05")

	level := formatLevel(r.Level)

	msg := r.Message

	file, line := getCaller()

	// final output
	fmt.Printf("%s %s %s (%s:%d)\n",
		colorGray(t),
		level,
		msg,
		file,
		line,
	)

	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return h
}

// 🎨 Colors + format

func formatLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return bold(colorBlue("[DEBUG]"))
	case slog.LevelInfo:
		return bold(colorGreen("[INFO]"))
	case slog.LevelWarn:
		return bold(colorYellow("[WARN]"))
	case slog.LevelError:
		return bold(colorRed("[ERROR]"))
	default:
		return bold("[HACK ]")
	}
}

func colorGray(s string) string   { return "\033[90m" + s + "\033[0m" }
func colorRed(s string) string    { return "\033[31m" + s + "\033[0m" }
func colorGreen(s string) string  { return "\033[32m" + s + "\033[0m" }
func colorYellow(s string) string { return "\033[33m" + s + "\033[0m" }
func colorBlue(s string) string   { return "\033[34m" + s + "\033[0m" }

func bold(s string) string { return "\033[1m" + s + "\033[0m" }

// ---- printf wrapper

func Infof(format string, args ...any) {
	Log.Info(fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...any) {
	Log.Debug(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	Log.Warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	Log.Error(fmt.Sprintf(format, args...))
}

// 📍 Caller (аналог define из C)
func getCaller() (string, int) {
	// 3 = глубина стека (подбирается)
	_, file, line, ok := runtime.Caller(5)
	if !ok {
		return "???", 0
	}

	// только имя файла
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]

	return file, line
}
