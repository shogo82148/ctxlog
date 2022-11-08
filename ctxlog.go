package ctxlog

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Level defines log levels.
type Level int

const (
	// LevelDebug defines debug log level.
	LevelDebug Level = iota

	// LevelInfo defines info log level.
	LevelInfo

	// LevelWarn defines warn log level.
	LevelWarn

	// LevelError defines error log level.
	LevelError

	// LevelFatal defines fatal log level.
	LevelFatal

	// LevelPanic defines panic log level.
	LevelPanic

	// LevelNo defines an absent log level.
	LevelNo

	// LevelDisabled disables the logger.
	LevelDisabled

	// LevelTrace defines trace log level.
	LevelTrace Level = -1
	// Values less than TraceLevel are handled as numbers.
)

type Logger struct {
	mu        sync.Mutex  // ensures atomic writes; protects the following fields
	prefix    string      // prefix on each line to identify the logger (but see Lmsgprefix)
	flag      int         // properties
	out       io.Writer   // for accumulating text to write
	isDiscard atomic.Bool // whether out == io.Discard
	level     Level
}

var std = New(os.Stderr, "", LstdFlags)

// Default returns the standard logger used by the package-level output functions.
func Default() *Logger { return std }

func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{
		out:    out,
		prefix: prefix,
		flag:   flag,
	}
}

func (l *Logger) Writer() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
	l.isDiscard.Store(w == io.Discard)
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) Level() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

type Fields map[string]any

type mergedFields struct {
	parent *mergedFields
	fields Fields
}

type ctxKey struct {
	name string
}

var keyFields = &ctxKey{"ctxlog"}

func With(parent context.Context, fields Fields) context.Context {
	return context.WithValue(parent, keyFields, &mergedFields{
		parent: contextFields(parent),
		fields: fields,
	})
}

func contextFields(ctx context.Context) *mergedFields {
	f := ctx.Value(keyFields)
	if f == nil {
		return nil
	}
	return f.(*mergedFields)
}

func (f *mergedFields) merge(dest map[string]any) {
	if f.parent != nil {
		f.parent.merge(dest)
	}
	for k, v := range f.fields {
		dest[k] = v
	}
}

func (l *Logger) OutputContext(ctx context.Context, calldepth int, level Level, msg string, fields Fields) error {
	if level < l.Level() {
		return nil
	}

	now := time.Now().UTC() // get this early.

	// TODO: build the message

	// build the fields
	f := make(map[string]any)
	if parent := contextFields(ctx); parent != nil {
		parent.merge(fields)
	}
	for k, v := range fields {
		f[k] = v
	}

	if t, ok := f["time"]; ok {
		f["field.time"] = t
	}
	f["time"] = now

	if lv, ok := f["level"]; ok {
		f["level"] = lv
	}
	f["level"] = level

	if msg, ok := f["message"]; ok {
		f["field.message"] = msg
	}
	f["message"] = msg

	// TODO: cache buffer
	buf, err := json.Marshal(f)
	if err != nil {
		return err
	}
	buf = append(buf, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = l.out.Write(buf)
	return err
}

func (l *Logger) Trace(ctx context.Context, msg string, fields Fields) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(ctx, 2, LevelTrace, msg, fields)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields Fields) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(ctx, 2, LevelDebug, msg, fields)
}

func (l *Logger) Info(ctx context.Context, msg string, fields Fields) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(ctx, 2, LevelInfo, msg, fields)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields Fields) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(ctx, 2, LevelWarn, msg, fields)
}

func (l *Logger) Error(ctx context.Context, msg string, fields Fields) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(ctx, 2, LevelError, msg, fields)
}

func (l *Logger) FatalContext(ctx context.Context, msg string, fields Fields) {
	l.OutputContext(ctx, 2, LevelFatal, msg, fields)
	os.Exit(1)
}

func (l *Logger) PanicContext(ctx context.Context, msg string, fields Fields) {
	l.OutputContext(ctx, 2, LevelPanic, msg, fields)
	panic(msg)
}
