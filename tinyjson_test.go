package ctxlog

import (
	"math"
	"testing"
)

func TestAppendAny(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		// strings
		{
			in:   "abcdefghijklmnopqrstuvwxyz",
			want: `"abcdefghijklmnopqrstuvwxyz"`,
		},
		{
			in:   "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			want: `"ABCDEFGHIJKLMNOPQRSTUVWXYZ"`,
		},
		{
			in:   "<script>",
			want: `"\u003cscript\u003e"`,
		},
		{
			in:   "\\\"\n\r\t",
			want: `"\\\"\n\r\t"`,
		},
		{
			in:   "こんにちは世界",
			want: `"こんにちは世界"`,
		},
		{
			in:   "😎",
			want: `"😎"`,
		},
		{
			in:   "\u0000\u0001\u001a\u007f\u2028\u2029",
			want: `"\u0000\u0001\u001a\u007f\u2028\u2029"`,
		},
		{
			in:   "\x80",
			want: `"�"`,
		},

		// number
		{
			in:   42,
			want: `42`,
		},
		{
			in:   int8(42),
			want: `42`,
		},
		{
			in:   int16(math.MinInt16),
			want: `-32768`,
		},
		{
			in:   int32(math.MinInt32),
			want: `-2147483648`,
		},
		{
			in:   int64(math.MinInt64),
			want: `-9223372036854775808`,
		},
		{
			in:   1.0,
			want: `1`,
		},
		{
			in:   float64(1e21),
			want: `1e+21`,
		},
		{
			in:   math.Nextafter(float64(1e21), 0),
			want: `999999999999999900000`,
		},
		{
			in:   float64(1e-6),
			want: `0.000001`,
		},
		{
			in:   math.Nextafter(float64(1e-6), 0),
			want: `9.999999999999997e-7`,
		},
		{
			in:   float32(1e21),
			want: `1e+21`,
		},
		{
			in:   math.Nextafter32(float32(1e21), 0),
			want: `999999950000000000000`,
		},
		{
			in:   float32(1e-6),
			want: `0.000001`,
		},
		{
			in:   math.Nextafter32(float32(1e-6), 0),
			want: `9.999999e-7`,
		},
	}

	e := newEncodeState()
	for i, tt := range tests {
		e.Reset()
		e.appendAny(tt.in)

		got := e.String()
		if got != tt.want {
			t.Errorf("%d: got %q, want %q", i, got, tt.want)
		}
	}
}
