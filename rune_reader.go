package deb822

import (
	"bufio"
	"io"
	"unicode"
)

type RuneReader struct {
	*bufio.Reader
}

func NewRuneReader(r io.Reader) *RuneReader {
	return &RuneReader{Reader: bufio.NewReader(r)}
}

func (r *RuneReader) PeekRune() (rune, int, error) {
	rn, size, err := r.ReadRune()
	if err != nil {
		return rn, size, err
	}

	return rn, size, r.UnreadRune()
}

func (r *RuneReader) DiscardRune() {
	_, _, _ = r.ReadRune()
}

func (r *RuneReader) DiscardSpace() {
	for {
		peek, _, err := r.ReadRune()
		if err != nil || !unicode.IsSpace(peek) {
			_ = r.UnreadRune()
			return
		}
	}
}
