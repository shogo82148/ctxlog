package ctxlog

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		flag int
		now  time.Time
		want string
	}{
		{
			flag: Ldate | LUTC,
			now:  time.Date(2001, 2, 3, 4, 5, 6, 123456789, time.UTC),
			want: "2001-02-03Z",
		},
		{
			flag: Ldate | Ltime | LUTC,
			now:  time.Date(2001, 2, 3, 4, 5, 6, 123456789, time.UTC),
			want: "2001-02-03T04:05:06Z",
		},
		{
			flag: Ldate | Lmicroseconds | LUTC,
			now:  time.Date(2001, 2, 3, 4, 5, 6, 123456789, time.UTC),
			want: "2001-02-03T04:05:06.123456Z",
		},
	}

	for i, tt := range tests {
		e := new(encodeState)
		e.appendTime(tt.flag, tt.now)
		got := e.String()
		if got != tt.want {
			t.Errorf("%d: got %q, want %q", i, got, tt.want)
		}
	}
}

func TestOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf, "", LstdFlags)
	l.Printf("hello %d world", 23)

	var got struct {
		Time    string
		Message string
		Level   string
	}
	t.Log(buf.String())
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
}

func TestOutputFields(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf, "", LstdFlags)

	parent := map[string]any{
		"parent": "hello",
	}
	ctx := With(context.Background(), parent)
	fields := map[string]any{
		"string":  "foobar",
		"number":  42,
		"boolean": true,
		"time":    "reserved",
		"message": "reserved",
		"file":    "reserved",
		"line":    "reserved",
	}
	l.Info(ctx, "hoge", fields)

	var got struct {
		Time    string
		Message string
		Level   string
	}
	t.Log(buf.String())
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Message != "hoge" {
		t.Errorf("got %q, want %q", got.Message, "hoge")
	}
}

func TestStackTrace(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf, "", Lshortfile)
	l.Print("hello")

	var got struct {
		Time    string
		Message string
		Level   string
		File    string
		Line    int
	}
	t.Log(buf.String())
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.File != "ctxlog_test.go" {
		t.Errorf("unexpected file name: got %q, want \"ctxlog_test.go\"", got.File)
	}
	if got.Line == 0 {
		t.Errorf("unexpected line number: %d", got.Line)
	}
}

type blackhole struct{}

// discard is same as io.Discard, but it avoids optimization to io.Discard.
var discard = &blackhole{}

func (b *blackhole) Write(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkStackTrace(b *testing.B) {
	const testString = "test"
	l := New(discard, "", Lshortfile)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Print(testString)
		}
	})
}

func BenchmarkFormatTime(b *testing.B) {
	e := new(encodeState)
	now := time.Date(2001, 2, 3, 4, 5, 6, 123456789, time.UTC)
	for i := 0; i < b.N; i++ {
		e.Reset()
		e.appendTime(Ldate|Ltime|Lmicroseconds|LUTC, now)
	}
}

func BenchmarkPrintln(b *testing.B) {
	const testString = "test"
	l := New(discard, "", LstdFlags)
	for i := 0; i < b.N; i++ {
		l.Println(testString)
	}
}

func BenchmarkPrintlnNoFlags(b *testing.B) {
	const testString = "test"
	l := New(discard, "", 0)
	for i := 0; i < b.N; i++ {
		l.Println(testString)
	}
}

func BenchmarkOutputFlag(b *testing.B) {
	parent := map[string]any{
		"parent": "hello",
	}
	ctx := With(context.Background(), parent)
	fields := map[string]any{
		"string":  "foobar",
		"number":  42,
		"boolean": true,
	}

	const testString = "test"
	l := New(discard, "", LstdFlags)
	for i := 0; i < b.N; i++ {
		l.Info(ctx, testString, fields)
	}
}

func BenchmarkOutputFlagParallel(b *testing.B) {
	parent := map[string]any{
		"parent": "hello",
	}
	ctx := With(context.Background(), parent)
	fields := map[string]any{
		"string":  "foobar",
		"number":  42,
		"boolean": true,
	}

	const testString = "test"
	l := New(discard, "", LstdFlags)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info(ctx, testString, fields)
		}
	})
}
