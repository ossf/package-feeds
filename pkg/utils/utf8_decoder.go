package utils

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"
)

// UTF8OnlyReader implements a Reader which ignores non utf-8 characters.
// Optionally, it can also replace non-XML characters with U+25A1 (□).
type UTF8OnlyReader struct {
	buffer             *bufio.Reader
	replaceNonXMLChars bool
}

// NewUTF8OnlyReader wraps a new UTF8OnlyReader around an existing reader.
// If replaceNonXMLChars is true, the reader will replace all invalid XML
// characters with the unicode replacement character.
func NewUTF8OnlyReader(rd io.Reader, replaceNonXMLChars bool) UTF8OnlyReader {
	return UTF8OnlyReader{bufio.NewReader(rd), replaceNonXMLChars}
}

// Returns true iff the given rune is in the XML Character Range, as defined
// by https://www.xml.com/axml/testaxml.htm, Section 2.2 Characters.
// Implementation copied from xml/xml.go.
func isInCharacterRange(r rune) bool {
	return r == 0x09 || r == 0x0A || r == 0x0D ||
		r >= 0x20 && r <= 0xD7FF ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}

// Reads bytes into the byte array b whilst ignoring non utf-8 characters
// Returns the number of bytes read and an error if one occurs.
func (reader UTF8OnlyReader) Read(b []byte) (int, error) {
	numBytesRead := 0
	for {
		r, runeSize, err := reader.buffer.ReadRune()
		if err != nil {
			return numBytesRead, err
		}

		// Invalid UTF-8 characters are represented with r set to utf8.RuneError
		// (i.e. Unicode replacement character) and read size of 1
		if r == utf8.RuneError && runeSize == 1 {
			continue
		}

		// Also ignore the replacement character for compatibility with previous behaviour
		// (yes, utf8.RuneError == unicode.ReplacementChar)
		if r == unicode.ReplacementChar {
			continue
		}

		if reader.replaceNonXMLChars && !isInCharacterRange(r) {
			r = '\u25A1' // □ (symbol for missing character)
			runeSize = utf8.RuneLen(r)
		}

		// Finish Read if we don't have enough space for this rune in the output byte slice
		if numBytesRead+runeSize >= len(b) {
			err = reader.buffer.UnreadRune()
			return numBytesRead, err
		}

		utf8.EncodeRune(b[numBytesRead:], r)
		numBytesRead += runeSize
	}
}
