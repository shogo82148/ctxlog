package ctxlog

import (
	"context"
	"fmt"
	"io"
	"os"
)

// compatible layer for the log package

const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix                    // move the "prefix" from the beginning of the line to before the message
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

func (l *Logger) Output(calldepth int, s string) error {
	return l.OutputContext(context.Background(), calldepth+1, LevelNo, s, nil)
}

// Print calls l.OutputContext to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...any) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(context.Background(), 2, LevelNo, fmt.Sprint(v...), nil)
}

// Printf calls l.OutputContext to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...any) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(context.Background(), 2, LevelNo, fmt.Sprintf(format, v...), nil)
}

// Println calls l.OutputContext to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Println(v ...any) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(context.Background(), 2, LevelNo, fmt.Sprint(v...), nil)
}

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func (l *Logger) Fatal(v ...any) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(context.Background(), 2, LevelFatal, fmt.Sprint(v...), nil)
	os.Exit(1)
}

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func (l *Logger) Fatalf(format string, v ...any) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(context.Background(), 2, LevelFatal, fmt.Sprintf(format, v...), nil)
	os.Exit(1)
}

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func (l *Logger) Fatalln(v ...any) {
	if l.isDiscard.Load() {
		return
	}
	l.OutputContext(context.Background(), 2, LevelFatal, fmt.Sprint(v...), nil)
	os.Exit(1)
}

// Panic is equivalent to l.Print() followed by a call to panic().
func (l *Logger) Panic(v ...any) {
	s := fmt.Sprint(v...)
	l.OutputContext(context.Background(), 2, LevelPanic, s, nil)
	panic(s)
}

// Panicf is equivalent to l.Printf() followed by a call to panic().
func (l *Logger) Panicf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	l.OutputContext(context.Background(), 2, LevelPanic, s, nil)
	panic(s)
}

func (l *Logger) Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	l.OutputContext(context.Background(), 2, LevelPanic, s, nil)
	panic(s)
}

func (l *Logger) Prefix() string {
	// TODO: implement me
	return ""
}

func (l *Logger) SetPrefix(prefix string) {
	// TODO: implement me
}

func (l *Logger) Flags() int {
	// TODO: implement me
	return 0
}

func (l *Logger) SetFlags(flag int) {
	// TODO: implement me
}

// Output writes the output for a logging event.
func Output(calldepth int, s string) error {
	return std.OutputContext(context.Background(), calldepth+1, LevelNo, s, nil) // +1 for this frame.
}

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...any) {
	if std.isDiscard.Load() {
		return
	}
	std.OutputContext(context.Background(), 2, LevelNo, fmt.Sprint(v...), nil)
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...any) {
	if std.isDiscard.Load() {
		return
	}
	std.OutputContext(context.Background(), 2, LevelNo, fmt.Sprintf(format, v...), nil)
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...any) {
	if std.isDiscard.Load() {
		return
	}
	std.OutputContext(context.Background(), 2, LevelNo, fmt.Sprintln(v...), nil)
}

func Fatal(v ...any) {
	std.OutputContext(context.Background(), 2, LevelFatal, fmt.Sprint(v...), nil)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	std.OutputContext(context.Background(), 2, LevelFatal, fmt.Sprintf(format, v...), nil)
	os.Exit(1)
}

func Fatalln(v ...any) {
	std.OutputContext(context.Background(), 2, LevelFatal, fmt.Sprint(v...), nil)
	os.Exit(1)
}

func Panic(v ...any) {
	s := fmt.Sprint(v...)
	std.OutputContext(context.Background(), 2, LevelPanic, s, nil)
	panic(s)
}

func Panicf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	std.OutputContext(context.Background(), 2, LevelPanic, s, nil)
	panic(s)
}

func Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	std.OutputContext(context.Background(), 2, LevelPanic, s, nil)
	panic(s)
}

func Prefix() string {
	return std.Prefix()
}

func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}

func Flags() int {
	return std.Flags()
}

func SetFlags(flag int) {
	std.SetFlags(flag)
}

func Writer() io.Writer {
	return std.Writer()
}

func SetOutput(w io.Writer) {
	std.SetOutput(w)
}
