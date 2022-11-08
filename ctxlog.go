package ctxlog

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"runtime"
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

func (lv Level) String() string {
	switch lv {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelPanic:
		return "panic"
	case LevelNo:
		return "no"
	case LevelDisabled:
		return "disabled"
	}
	return "trace"
}

type Logger struct {
	mu        sync.RWMutex // ensures atomic writes; protects the following fields
	prefix    string       // prefix on each line to identify the logger (but see Lmsgprefix)
	flag      int          // properties
	out       io.Writer    // for accumulating text to write
	isDiscard atomic.Bool  // whether out == io.Discard
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

	now := time.Now() // get this early.

	// TODO: build the message

	// build the fields
	f := make(map[string]any)
	if parent := contextFields(ctx); parent != nil {
		parent.merge(f)
	}
	for k, v := range fields {
		f[k] = v
	}

	if t, ok := f["time"]; ok {
		f["field.time"] = t
	}
	f["time"] = l.formatTime(now)

	if lv, ok := f["level"]; ok {
		f["level"] = lv
	}
	f["level"] = level.String()

	if msg, ok := f["message"]; ok {
		f["field.message"] = msg
	}
	if l.Flags()&Lmsgprefix == 0 {
		msg = l.Prefix() + msg
	} else {
		msg = msg + l.Prefix()
	}
	f["message"] = msg

	// stack trace
	if l.Flags()&(Lshortfile|Llongfile) != 0 {
		_, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		} else {
			if l.flag&Lshortfile != 0 {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						short = file[i+1:]
						break
					}
				}
				file = short
			}
		}
		if v, ok := f["file"]; ok {
			f["field.file"] = v
		}
		if v, ok := f["line"]; ok {
			f["field.line"] = v
		}
		f["file"] = file
		f["line"] = line
	}

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

func (l *Logger) formatTime(t time.Time) string {
	var buf [30]byte
	var idx int

	flag := l.Flags()
	if flag&LUTC != 0 {
		t = t.UTC()
	}
	if flag&Ldate != 0 {
		year, month, day := t.Date()
		buf[idx+0] = '0' + byte(year/1000)
		buf[idx+1] = '0' + byte((year/100)%10)
		buf[idx+2] = '0' + byte((year/10)%10)
		buf[idx+3] = '0' + byte(year%10)
		buf[idx+4] = '-'
		buf[idx+5] = '0' + byte(month/10)
		buf[idx+6] = '0' + byte(month%10)
		buf[idx+7] = '-'
		buf[idx+8] = '0' + byte(day/10)
		buf[idx+9] = '0' + byte(day%10)
		idx += 10
	}
	if flag&(Ltime|Lmicroseconds) != 0 {
		if l.flag&Ldate != 0 {
			buf[idx] = 'T'
			idx++
		}
		hour, min, sec := t.Clock()
		buf[idx+0] = '0' + byte(hour/10)
		buf[idx+1] = '0' + byte(hour%10)
		buf[idx+2] = ':'
		buf[idx+3] = '0' + byte(min/10)
		buf[idx+4] = '0' + byte(min%10)
		buf[idx+5] = ':'
		buf[idx+6] = '0' + byte(sec/10)
		buf[idx+7] = '0' + byte(sec%10)
		idx += 8
		if flag&(Lmicroseconds) != 0 {
			micro := t.Nanosecond() / 1000
			buf[idx+0] = '.'
			buf[idx+1] = '0' + byte(micro/100000)
			buf[idx+2] = '0' + byte((micro/10000)%10)
			buf[idx+3] = '0' + byte((micro/1000)%10)
			buf[idx+4] = '0' + byte((micro/100)%10)
			buf[idx+5] = '0' + byte((micro/10)%10)
			buf[idx+6] = '0' + byte(micro%10)
			idx += 7
		}
	}
	if flag&LUTC != 0 {
		buf[idx] = 'Z'
		idx++
	}
	return string(buf[:idx])
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

func Trace(ctx context.Context, msg string, fields Fields) {
	if std.isDiscard.Load() {
		return
	}
	std.OutputContext(ctx, 2, LevelTrace, msg, fields)
}

func Debug(ctx context.Context, msg string, fields Fields) {
	if std.isDiscard.Load() {
		return
	}
	std.OutputContext(ctx, 2, LevelDebug, msg, fields)
}

func Info(ctx context.Context, msg string, fields Fields) {
	if std.isDiscard.Load() {
		return
	}
	std.OutputContext(ctx, 2, LevelInfo, msg, fields)
}

func Warn(ctx context.Context, msg string, fields Fields) {
	if std.isDiscard.Load() {
		return
	}
	std.OutputContext(ctx, 2, LevelWarn, msg, fields)
}

func Error(ctx context.Context, msg string, fields Fields) {
	if std.isDiscard.Load() {
		return
	}
	std.OutputContext(ctx, 2, LevelError, msg, fields)
}

func FatalContext(ctx context.Context, msg string, fields Fields) {
	std.OutputContext(ctx, 2, LevelFatal, msg, fields)
	os.Exit(1)
}

func PanicContext(ctx context.Context, msg string, fields Fields) {
	std.OutputContext(ctx, 2, LevelPanic, msg, fields)
	panic(msg)
}
