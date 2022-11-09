package ctxlog

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
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
	}

	e := newEncodeState()
	for i, tt := range tests {
		e.Reset()
		e.appendString(tt.in)

		got := e.String()
		if got != tt.want {
			t.Errorf("%d: got %q, want %q", i, got, tt.want)
		}
	}
}
