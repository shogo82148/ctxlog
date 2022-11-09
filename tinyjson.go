package ctxlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"
)

var reservedFields = []string{
	"time",
	"level",
	"file",
	"line",
	"message",
}

type keyValue struct {
	key   string
	value any
}

type keyValues []keyValue

var _ sort.Interface = keyValues(nil)

func (s keyValues) Len() int {
	return len(s)
}

func (s keyValues) Less(i, j int) bool {
	return s[i].key < s[j].key
}

func (s keyValues) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type encodeState struct {
	bytes.Buffer // accumulated output
	scratch      [64]byte
	kv           []keyValue
	enc          *json.Encoder
}

func newEncodeState() *encodeState {
	e := new(encodeState)
	e.enc = json.NewEncoder(&e.Buffer)
	return e
}

func (e *encodeState) appendRawString(v string) {
	const hex = "0123456789abcdef"
	for _, c := range v {
		switch c {
		case '\\':
			e.WriteString(`\\`)
		case '"':
			e.WriteString(`\"`)
		case '\n':
			e.WriteString(`\n`)
		case '\r':
			e.WriteString(`\r`)
		case '\t':
			e.WriteString(`\t`)
		case '<':
			e.WriteString(`\u003c`)
		case '>':
			e.WriteString(`\u003e`)
		case '&':
			e.WriteString(`\u0026`)
		case '\u007f':
			e.WriteString(`\u007f`)

		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		case '\u2028':
			e.WriteString(`\u2028`)
		case '\u2029':
			e.WriteString(`\u2029`)

		default:
			if c < 0x20 {
				// control characters
				e.WriteByte('\\')
				e.WriteByte('u')
				e.WriteByte('0')
				e.WriteByte('0')
				e.WriteByte(hex[(c/0x10)%0x10])
				e.WriteByte(hex[c%0x10])
			} else {
				e.WriteRune(c)
			}
		}
	}
}

func (e *encodeState) appendString(v string) {
	e.WriteByte('"')
	e.appendRawString(v)
	e.WriteByte('"')
}

func (e *encodeState) appendBool(v bool) {
	if v {
		e.WriteString("true")
	} else {
		e.WriteString("false")
	}
}

func (e *encodeState) appendInt(v int64) {
	b := strconv.AppendInt(e.scratch[:0], v, 10)
	e.Write(b)
}

func (e *encodeState) appendUint(v uint64) {
	b := strconv.AppendUint(e.scratch[:0], v, 10)
	e.Write(b)
}

func (e *encodeState) appendFloat64(v float64) error {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return fmt.Errorf("ctxio: unsupported value: %g", v)
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	abs := math.Abs(v)
	fmt := byte('f')
	if abs != 0 && (abs < 1e-6 || abs >= 1e21) {
		fmt = byte('e')
	}
	b := strconv.AppendFloat(e.scratch[:0], v, fmt, -1, 64)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	e.Write(b)
	return nil
}

func (e *encodeState) appendFloat32(v float32) error {
	f := float64(v)
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return fmt.Errorf("ctxio: unsupported value: %g", v)
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	abs := math.Abs(f)
	fmt := byte('f')
	if abs != 0 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
		fmt = byte('e')
	}
	b := strconv.AppendFloat(e.scratch[:0], f, fmt, -1, 32)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	e.Write(b)
	return nil
}

func (e *encodeState) appendTime(flags int, t time.Time) {
	b := &e.scratch
	var i int

	if flags&LUTC != 0 {
		t = t.UTC()
	}
	if flags&Ldate != 0 {
		year, month, day := t.Date()
		b[i+0] = '0' + byte(year/1000)
		b[i+1] = '0' + byte((year/100)%10)
		b[i+2] = '0' + byte((year/10)%10)
		b[i+3] = '0' + byte(year%10)
		b[i+4] = '-'
		b[i+5] = '0' + byte(month/10)
		b[i+6] = '0' + byte(month%10)
		b[i+7] = '-'
		b[i+8] = '0' + byte(day/10)
		b[i+9] = '0' + byte(day%10)
		i += 10
	}
	if flags&(Ltime|Lmicroseconds) != 0 {
		if flags&Ldate != 0 {
			b[i] = 'T'
			i++
		}
		hour, min, sec := t.Clock()
		b[i+0] = '0' + byte(hour/10)
		b[i+1] = '0' + byte(hour%10)
		b[i+2] = ':'
		b[i+3] = '0' + byte(min/10)
		b[i+4] = '0' + byte(min%10)
		b[i+5] = ':'
		b[i+6] = '0' + byte(sec/10)
		b[i+7] = '0' + byte(sec%10)
		i += 8
		if flags&(Lmicroseconds) != 0 {
			micro := t.Nanosecond() / 1000
			b[i+0] = '.'
			b[i+1] = '0' + byte(micro/100000)
			b[i+2] = '0' + byte((micro/10000)%10)
			b[i+3] = '0' + byte((micro/1000)%10)
			b[i+4] = '0' + byte((micro/100)%10)
			b[i+5] = '0' + byte((micro/10)%10)
			b[i+6] = '0' + byte(micro%10)
			i += 7
		}
	}
	if flags&LUTC != 0 {
		b[i] = 'Z'
		i++
	}
	e.Write((*b)[:i])
}

func (e *encodeState) appendAny(v any) error {
	switch v := v.(type) {
	case int8:
		e.appendInt(int64(v))
	case int16:
		e.appendInt(int64(v))
	case int32:
		e.appendInt(int64(v))
	case int64:
		e.appendInt(int64(v))
	case int:
		e.appendInt(int64(v))
	case uint8:
		e.appendUint(uint64(v))
	case uint16:
		e.appendUint(uint64(v))
	case uint32:
		e.appendUint(uint64(v))
	case uint64:
		e.appendUint(uint64(v))
	case uint:
		e.appendUint(uint64(v))
	case string:
		e.appendString(v)
	case bool:
		e.appendBool(v)
	case float32:
		return e.appendFloat32(v)
	case float64:
		return e.appendFloat64(v)
	default:
		if err := e.enc.Encode(v); err != nil {
			return err
		}
	}
	return nil
}

func (e *encodeState) appendFields(parent *mergedFields, fields Fields) error {
	kv := e.kv[:0]
	for k, v := range fields {
		kv = append(kv, keyValue{key: k, value: v})
	}
	for parent != nil {
		for k, v := range parent.fields {
			kv = append(kv, keyValue{key: k, value: v})
		}
		parent = parent.parent
	}
	sort.Stable(keyValues(kv))

	for i, pair := range kv {
		if i > 0 && kv[i-1].key == pair.key {
			continue
		}
		e.WriteByte(',')
		e.WriteByte('"')
		for _, k := range reservedFields {
			if pair.key == k {
				e.appendRawString("field.")
				break
			}
		}
		e.appendRawString(pair.key)
		e.WriteByte('"')
		e.WriteByte(':')
		if err := e.appendAny(pair.value); err != nil {
			return err
		}
	}

	// fill with nil for Garbage Collection
	for i := range kv {
		kv[i].key = ""
		kv[i].value = nil
	}
	e.kv = kv
	return nil
}
