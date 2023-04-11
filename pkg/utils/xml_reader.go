package utils

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"
)

const unicodeWhiteSquare = '\u25A1'

// XMLReader implements a Reader that reads XML responses and does two things:
// 1. Ignores non UTF-8 characters.
// 2. If ReplacementChar is not rune(0), non-XML valid runes are replaced with that.
type XMLReader struct {
	buffer          *bufio.Reader
	ReplacementChar rune
}

// NewXMLReader wraps a new XMLReader around an existing reader.
// replaceNonXMLChars is a convenience option that sets ReplacementChar
// to the unicode white square character. This character will replace
// all invalid XML characters found in the stream.
func NewXMLReader(rd io.Reader, replaceNonXMLChars bool) XMLReader {
	reader := XMLReader{buffer: bufio.NewReader(rd)}
	if replaceNonXMLChars {
		reader.ReplacementChar = unicodeWhiteSquare
	}
	return reader
}

func (reader XMLReader) replaceNonXMLChars() bool {
	return reader.ReplacementChar != rune(0)
}

// Returns true iff the given rune is in the XML Character Range, as defined
// by https://www.xml.com/axml/testaxml.htm, Section 2.2 Characters.
// Implementation copied from xml/xml.go.
func isInXMLCharacterRange(r rune) bool {
	return r == 0x09 || r == 0x0A || r == 0x0D ||
		r >= 0x20 && r <= 0xD7FF ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}

// Reads bytes into the byte array b whilst ignoring non utf-8 characters
// Returns the number of bytes read and an error if one occurs.
func (reader XMLReader) Read(b []byte) (int, error) {
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

		if reader.replaceNonXMLChars() && !isInXMLCharacterRange(r) {
			r = reader.ReplacementChar
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
