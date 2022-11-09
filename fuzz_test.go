package ctxlog

import (
	"encoding/json"
	"testing"
)

func FuzzString(f *testing.F) {
	f.Add("abcdefghijklmnopqrstuvwxyz")
	f.Add("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	f.Add("<script>")
	f.Add("\\\"\n\r\t")
	f.Add("こんにちは世界")
	f.Add("😎")

	f.Fuzz(func(t *testing.T, s string) {
		e := newEncodeState()
		e.appendString(s)

		data := e.Bytes()
		if !json.Valid(data) {
			t.Errorf("invalid json: %q", string(data))
		}
	})
}
