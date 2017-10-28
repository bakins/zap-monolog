// Package monolog wraps zap https://github.com/uber-go/zap
// to log similarly to https://github.com/Seldaek/monolog
// This is useful if you have a log pipeline that expects
// logs in monolog json format
package monolog

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ProcessorFunc is used to add extra fields.
type ProcessorFunc func() zapcore.Field

// Logger wraps a zap logger
type Logger struct {
	Logger     *zap.Logger
	Config     zap.Config
	Level      zap.AtomicLevel
	Processors []ProcessorFunc
}

// OptionsFunc is used to set options
type OptionsFunc func(*Logger) error

// New creates a new logger.
func New(options ...OptionsFunc) (*Logger, error) {
	a := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	l := &Logger{
		Config: zap.Config{
			Development:       false,
			DisableCaller:     true,
			DisableStacktrace: true,
			EncoderConfig:     zap.NewProductionEncoderConfig(),
			Encoding:          "json",
			ErrorOutputPaths:  []string{"stderr"},
			Level:             a,
			OutputPaths:       []string{"stdout"},
		},
		Level: a,
	}

	for _, f := range options {
		if err := f(l); err != nil {
			return nil, errors.Wrap(err, "options function failed")
		}
	}

	logger, err := l.Config.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build logger")
	}
	l.Logger = logger
	return l, nil
}

// With creates a child logger and adds structured context to it. Fields added to the child don't affect the parent, and vice versa.
// Processors are copied and changes to child do not  affect the parent, and vice versa.
// The level is shared between child and parent.
func (l *Logger) With(fields ...zapcore.Field) *Logger {
	n := &Logger{
		Logger:     l.Logger.With(fields...),
		Config:     l.Config,
		Level:      l.Level,
		Processors: make([]ProcessorFunc, len(l.Processors)),
	}

	for i, p := range l.Processors {
		n.Processors[i] = p
	}
	return n
}

type extraFields struct {
	fields []*zapcore.Field
}

func (e *extraFields) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, f := range e.fields {
		f.AddTo(enc)
	}
	return nil
}

func (l *Logger) write(level zapcore.Level, msg string, fields []zapcore.Field) {
	if !l.Level.Enabled(level) {
		return
	}

	extras := &extraFields{
		fields: make([]*zapcore.Field, len(l.Processors)),
	}

	i := 0
	for _, p := range l.Processors {
		f := p()
		if f.Type != zapcore.UnknownType {
			extras.fields[i] = &f
			i++
		}
	}

	ctx := make([]zapcore.Field, len(fields)+2)
	ctx[0] = zap.Object("extra", extras)
	ctx[1] = zap.Namespace("context")

	i = 2
	for _, f := range fields {
		if f.Type != zapcore.UnknownType {
			ctx[i] = f
			i++
		}
	}

	l.Logger.Info(msg, ctx...)
}

// Info logs a message at InfoLevel.
func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	l.write(zapcore.InfoLevel, msg, fields)
}
