package ctxlog

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"
)

type encodeState struct {
	bytes.Buffer // accumulated output
	scratch      [64]byte
}

func (e *encodeState) appendRawString(v string) {
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
			e.WriteRune(c)
		}
	}
}

func (e *encodeState) appendString(v string) {
	e.WriteByte('"')
	e.appendRawString(v)
	e.WriteByte('"')
}

func (e *encodeState) appendInt(v int64) {
	b := strconv.AppendInt(e.scratch[:0], v, 10)
	e.Write(b)
}

func (e *encodeState) appendUint(v uint64) {
	b := strconv.AppendUint(e.scratch[:0], v, 10)
	e.Write(b)
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

func (e *encodeState) appendAny(buf *bytes.Buffer, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	buf.Write(b)
	return nil
}
