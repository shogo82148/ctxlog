package ctxlog

import (
	"bytes"
	"context"
	"testing"
)

func BenchmarkPrintln(b *testing.B) {
	const testString = "test"
	var buf bytes.Buffer
	l := New(&buf, "", LstdFlags)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Println(testString)
	}
}

func BenchmarkPrintlnNoFlags(b *testing.B) {
	const testString = "test"
	var buf bytes.Buffer
	l := New(&buf, "", 0)
	for i := 0; i < b.N; i++ {
		buf.Reset()
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
	var buf bytes.Buffer
	l := New(&buf, "", LstdFlags)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Info(ctx, testString, fields)
	}
}
