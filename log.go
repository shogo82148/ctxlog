package ctxlog

import (
	"context"
	"io"
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

func (l *Logger) Print(v ...any) {
	if l.isDiscard.Load() {
		return
	}
	// TODO: implement me
}

func (l *Logger) Printf(format string, v ...any) {
	// TODO: implement me
}

func (l *Logger) Println(v ...any) {
	// TODO: implement me
}

func (l *Logger) Fatal(v ...any) {
	// TODO: implement me
}

func (l *Logger) Fatalf(format string, v ...any) {
	// TODO: implement me
}

func (l *Logger) Fatalln(v ...any) {
	// TODO: implement me
}

func (l *Logger) Panic(v ...any) {
	// TODO: implement me
}

func (l *Logger) Panicf(format string, v ...any) {
	// TODO: implement me
}

func (l *Logger) Panicln(v ...any) {
	// TODO: implement me
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

func Output(calldepth int, s string) error {
	// TODO: implement me
	return nil
}

func Print(v ...any) {
	// TODO: implement me
}

func Printf(format string, v ...any) {
	// TODO: implement me
}

func Println(v ...any) {
	// TODO: implement me
}

func Fatal(v ...any) {
	// TODO: implement me
}

func Fatalf(format string, v ...any) {
	// TODO: implement me
}

func Fatalln(v ...any) {
	// TODO: implement me
}

func Panic(v ...any) {
	// TODO: implement me
}

func Panicf(format string, v ...any) {
	// TODO: implement me
}

func Panicln(v ...any) {
	// TODO: implement me
}

func Prefix() string {
	// TODO: implement me
	return ""
}

func SetPrefix(prefix string) {
	// TODO: implement me
}

func Flags() int {
	// TODO: implement me
	return 0
}

func SetFlags(flag int) {
	// TODO: implement me
}

func Writer() io.Writer {
	// TODO: implement me
	return nil
}

func SetOutput(w io.Writer) {
	// TODO: implement me
}
