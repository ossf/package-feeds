package utils

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"
)

// ValidUTF8Reader implements a Reader which ignores non utf-8 characters.
type UTF8OnlyReader struct {
	buffer *bufio.Reader
}

// NewValidUTF8Reader wraps a new UTF8OnlyReader around an existing reader.
func NewUTF8OnlyReader(rd io.Reader) UTF8OnlyReader {
	return UTF8OnlyReader{bufio.NewReader(rd)}
}

// Reads bytes into the byte array b whilst ignoring non utf-8 characters
// Returns the number of bytes read and an error if one occurs.
func (reader UTF8OnlyReader) Read(b []byte) (int, error) {
	var numBytesRead int
	for {
		r, runeSize, err := reader.buffer.ReadRune()
		if err != nil {
			return numBytesRead, err
		}

		// Ignore non utf-8 characters
		if r == unicode.ReplacementChar {
			continue
		}

		// Finish Read if we don't have enough space for this rune in the output byte slice
		if len(b)-numBytesRead <= runeSize {
			err = reader.buffer.UnreadRune()
			return numBytesRead, err
		}

		utf8.EncodeRune(b[numBytesRead:], r)
		numBytesRead += runeSize
	}
}
