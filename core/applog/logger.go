package applog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	DefaultMaxFileBytes int64 = 1 << 20 // 1MB
	activeLogName             = "skillflow.log"
	backupLogName             = "skillflow.log.1"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

func ParseLevel(level string) Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return LevelDebug
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelError:
		return "error"
	default:
		return "info"
	}
}

type Logger struct {
	mu        sync.Mutex
	dir       string
	threshold Level
	maxBytes  int64
}

func New(dir, level string) (*Logger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Logger{
		dir:       dir,
		threshold: ParseLevel(level),
		maxBytes:  DefaultMaxFileBytes,
	}, nil
}

func (l *Logger) Dir() string {
	return l.dir
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.threshold = level
	l.mu.Unlock()
}

func (l *Logger) SetLevelString(level string) {
	l.SetLevel(ParseLevel(level))
}

func (l *Logger) LevelString() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.threshold.String()
}

func (l *Logger) Enabled(level Level) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return level >= l.threshold
}

func (l *Logger) Debugf(format string, args ...any) {
	l.Logf(LevelDebug, format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.Logf(LevelInfo, format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Logf(LevelError, format, args...)
}

func (l *Logger) Logf(level Level, format string, args ...any) {
	l.log(level, fmt.Sprintf(format, args...))
}

func (l *Logger) log(level Level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level < l.threshold {
		return
	}

	line := fmt.Sprintf("%s [%s] %s\n",
		time.Now().Format("2006-01-02 15:04:05.000"),
		strings.ToUpper(level.String()),
		message,
	)

	if err := l.rotateIfNeeded(int64(len(line))); err != nil {
		return
	}
	file, err := os.OpenFile(filepath.Join(l.dir, activeLogName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(line)
}

func (l *Logger) rotateIfNeeded(nextLineBytes int64) error {
	activePath := filepath.Join(l.dir, activeLogName)
	stat, err := os.Stat(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if stat.Size()+nextLineBytes <= l.maxBytes {
		return nil
	}

	backupPath := filepath.Join(l.dir, backupLogName)
	_ = os.Remove(backupPath)
	if err := os.Rename(activePath, backupPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
