// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package contents

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// maxLineLength bounds a single entry. A path shared by a few hundred
	// packages runs into the low kilobytes; a megabyte is well past anything a
	// real archive emits, and keeps a corrupt stream from being buffered whole.
	maxLineLength = 1 << 20

	// headerScanLimit bounds how far into the file the legacy header terminator
	// is looked for. Old Contents files opened with a paragraph of prose closed
	// by a "FILE LOCATION" column heading; the prose was never more than a
	// couple of dozen lines.
	headerScanLimit = 100
)

// line is a scanned line together with its 1-based number, kept so that errors
// can name the offending line even after the header lookahead replays it.
type line struct {
	no   int
	text string
}

// Reader reads a Contents index one entry at a time. The legacy prose header,
// if present, is skipped. Blank lines are ignored.
type Reader struct {
	s          *bufio.Scanner
	no         int
	buf        []line
	header     []string
	headerRead bool
}

// NewReader returns a Reader reading the uncompressed Contents index in r.
func NewReader(r io.Reader) *Reader {
	s := bufio.NewScanner(r)
	s.Buffer(nil, maxLineLength)

	return &Reader{s: s}
}

// Read returns the next entry. It returns io.EOF when the index is exhausted.
func (r *Reader) Read() (Entry, error) {
	if !r.headerRead {
		if err := r.readHeader(); err != nil {
			return Entry{}, err
		}
	}

	for {
		l, ok := r.next()
		if !ok {
			if err := r.s.Err(); err != nil {
				return Entry{}, fmt.Errorf("failed to read contents index: %w", err)
			}

			return Entry{}, io.EOF
		}

		text := strings.TrimRight(l.text, " \t\r")
		if text == "" {
			continue
		}

		entry, ok := parseEntry(text)
		if !ok {
			return Entry{}, fmt.Errorf("line %d: %w: %q", l.no, ErrInvalidEntry, l.text)
		}

		return entry, nil
	}
}

// ReadAll reads the remaining entries of the index.
func (r *Reader) ReadAll() ([]Entry, error) {
	var entries []Entry
	for {
		entry, err := r.Read()
		if errors.Is(err, io.EOF) {
			return entries, nil
		} else if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}
}

// Header returns the lines of the legacy prose header, without the "FILE
// LOCATION" line that closes it. It is empty for the headerless files every
// current generator writes, and is only populated once reading has started.
func (r *Reader) Header() []string {
	return r.header
}

// readHeader consumes a leading prose header, if one is there. Lines are
// buffered until the "FILE LOCATION" column heading turns up; if it does not
// turn up within headerScanLimit lines the file has no header and the buffered
// lines are replayed as entries.
func (r *Reader) readHeader() error {
	r.headerRead = true

	for range headerScanLimit {
		l, ok := r.scan()
		if !ok {
			return nil
		}

		if isHeaderTerminator(l.text) {
			for _, buffered := range r.buf {
				r.header = append(r.header, buffered.text)
			}
			r.buf = nil

			return nil
		}

		r.buf = append(r.buf, l)
	}

	return nil
}

// next returns the next line, draining the header lookahead buffer first.
func (r *Reader) next() (line, bool) {
	if len(r.buf) > 0 {
		l := r.buf[0]
		r.buf = r.buf[1:]

		return l, true
	}

	return r.scan()
}

// scan pulls one line straight off the underlying scanner.
func (r *Reader) scan() (line, bool) {
	if !r.s.Scan() {
		return line{}, false
	}
	r.no++

	return line{no: r.no, text: r.s.Text()}, true
}

// isHeaderTerminator reports whether the line is the "FILE LOCATION" column
// heading that closes a legacy header. Package names are lowercase, so no real
// entry collides with it.
func isHeaderTerminator(text string) bool {
	fields := strings.Fields(text)

	return len(fields) == 2 && fields[0] == "FILE" && fields[1] == "LOCATION"
}

// parseEntry splits a line into its two columns. The split is made on the last
// whitespace run of the line: the separator has no fixed width, and paths may
// contain spaces of their own.
func parseEntry(text string) (Entry, bool) {
	i := strings.LastIndexAny(text, " \t")
	if i < 0 {
		return Entry{}, false
	}

	path := strings.TrimRight(text[:i], " \t")
	if path == "" {
		return Entry{}, false
	}

	var packages []string
	for name := range strings.SplitSeq(text[i+1:], ",") {
		if name = strings.Trim(name, " \t"); name != "" {
			packages = append(packages, name)
		}
	}

	if len(packages) == 0 {
		return Entry{}, false
	}

	return Entry{Path: path, Packages: packages}, true
}
