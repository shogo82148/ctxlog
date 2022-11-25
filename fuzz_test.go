package ctxlog

import (
	"encoding/json"
	"reflect"
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

func FuzzTinyJSON(f *testing.F) {
	f.Add(`{}`, `{}`)
	f.Add(`{"foo":"bar"}`, `{"hoge":"fuga"}`)
	f.Add(`{"foo":"bar"}`, `{"foo":"fuga"}`)
	f.Add(`{"foo":"bar"}`, `{"hoge":1}`)
	f.Add(`{"foo":"bar"}`, `{"hoge":-1}`)
	f.Add(`{"foo":"bar"}`, `{"hoge":1.234e10}`)
	f.Add(`{"foo":"bar"}`, `{"hoge":true}`)
	f.Add(`{"foo":"bar"}`, `{"hoge":false}`)
	f.Add(`{"foo":"[2,3,4]]"}`, `{"foo":[1,2,3]}`)
	f.Add(`{"foo":"[2,3,4]]"}`, `{"foo":{"bar":"fiuzz"}}}`)

	f.Fuzz(func(t *testing.T, raw0, raw1 string) {
		var parent, child map[string]any
		if err := json.Unmarshal([]byte(raw0), &parent); err != nil {
			return
		}
		if err := json.Unmarshal([]byte(raw1), &child); err != nil {
			return
		}

		merged := make(map[string]any, len(parent)+len(child))
		for k, v := range parent {
			merged[k] = v
		}
		for k, v := range child {
			merged[k] = v
		}
		merged["message"] = ""

		e := newEncodeState()
		e.WriteString(`{"message":""`)
		if err := e.appendFields(&mergedFields{fields: Fields(parent)}, Fields(child)); err != nil {
			t.Fatal(err)
		}
		e.WriteByte('}')
		data := e.Bytes()

		var got map[string]any
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("failed to parse %q: %v", string(data), err)
		}
		if !reflect.DeepEqual(got, merged) {
			t.Errorf("got %#v, want %#v", got, merged)
		}
	})
}
