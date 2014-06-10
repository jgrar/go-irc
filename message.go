package irc

import (
	"bytes"
	"strings"
	"encoding/json"
	"errors"
	//"log"
)

type Message struct{
	Prefix, Command string
	Parameters []string
	Trailing string
}

func (m *Message) Unmarshal (text []byte) error {

	var i int
	if len(text) == 0 || text == nil {
		return nil
	}

	if text[0] == ':' {
		i = bytes.IndexByte(text, ' ')

		if i == -1 {
			return errors.New("Command missing")
		}

		m.Prefix = string(text[1:i])
		text = text[i + 1:]
	}

	i = bytes.IndexByte(text, ' ')

	if i == -1 {
		m.Command = string(text)
		return nil
	} else if text[0] == ':' {
		m.Trailing = string(text[1:])
		return errors.New("Command missing")
	}

	m.Command = string(text[:i])

	for {
		text = text[i + 1:]

		if text[0] == ':' {
			m.Trailing = string(text[1:])
			break
		}

		i = bytes.IndexByte(text, ' ')

		if i == -1 {
			m.Parameters = append(m.Parameters, string(text))
			break
		}

		m.Parameters = append(m.Parameters, string(text[:i]))
	}

	return nil
}

func (m *Message) Marshal () (text []byte, err error) {
	var buf bytes.Buffer

	if len(m.Prefix) != 0 {
		buf.WriteString(":" + m.Prefix + " ")
	}

	if len(m.Command) != 0 {
		buf.WriteString(m.Command)
	} else {
		err = errors.New("Command missing")
	}

	if len(m.Parameters) != 0 {
		buf.WriteString(" " + strings.Join(m.Parameters, " "))
	}

	if len(m.Trailing) != 0 {
		buf.WriteString(" :" + m.Trailing)
	}

	return buf.Bytes(), err
}

func ScanLines (data []byte, atEOF bool) (advance int, token []byte, err error) {

	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		return i + 2, data[:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil

}

func ScanJson (data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = ScanLines(data, atEOF)

	if err == nil && token != nil {
		msg := &Message{}
		err = msg.Unmarshal(token)

		if err == nil {
			token, err = json.Marshal(msg)
		}
	}

	return
}
