package deb822

import (
	"bufio"
	"io"
	"unicode"
)

type RuneReader struct {
	*bufio.Reader
}

// NewRuneReader creates a new RuneReader from the provided io.Reader.
func NewRuneReader(r io.Reader) *RuneReader {
	return &RuneReader{Reader: bufio.NewReader(r)}
}

// PeekRune peeks at the next rune without consuming it.
func (r *RuneReader) PeekRune() (rune, int, error) {
	rn, size, err := r.ReadRune()
	if err != nil {
		return rn, size, err
	}

	return rn, size, r.UnreadRune()
}

// DiscardRune discards the next rune.
func (r *RuneReader) DiscardRune() {
	_, _, _ = r.ReadRune()
}

// DiscardSpace discards all consecutive whitespace runes.
func (r *RuneReader) DiscardSpace() {
	for {
		peek, _, err := r.ReadRune()
		if err != nil || !unicode.IsSpace(peek) {
			_ = r.UnreadRune()
			return
		}
	}
}
