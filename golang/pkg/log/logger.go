package log

import "github.com/sirupsen/logrus"
import proxiedLog "github.com/sirupsen/logrus"

// Level type represents a logging level
type Level uint8

// Logging levels
const (
	LevelFatal Level = iota
	LevelError
	LevelWarning
	LevelInfo
)

// Format type represents a logging format
type Format uint8

// Format types
const (
	FormatHuman Format = iota
	FormatJSON
)

func init() {
	SetLevel(LevelError)
	SetFormat(FormatHuman)
}

// SetLevel sets the logging level
func SetLevel(level Level) {
	switch level {
	case LevelFatal:
		logrus.SetLevel(logrus.FatalLevel)
	case LevelInfo:
		logrus.SetLevel(logrus.InfoLevel)
	case LevelWarning:
		logrus.SetLevel(logrus.WarnLevel)
	case LevelError:
		logrus.SetLevel(logrus.ErrorLevel)
	}
}

// SetFormat sets the log format type
func SetFormat(format Format) {
	switch format {
	case FormatHuman:
		fmtter := logrus.TextFormatter{}
		fmtter.ForceColors = true // Force human (non logfmt) output on Windows
		logrus.SetFormatter(&fmtter)
	case FormatJSON:
		fmtter := logrus.JSONFormatter{}
		logrus.SetFormatter(&fmtter)
	}
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	proxiedLog.Fatal(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	proxiedLog.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	proxiedLog.Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(args ...interface{}) {
	proxiedLog.Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	proxiedLog.Error(args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	proxiedLog.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	proxiedLog.Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(format string, args ...interface{}) {
	proxiedLog.Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	proxiedLog.Errorf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	proxiedLog.Fatalf(format, args...)
}
