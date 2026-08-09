// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package changelog

import (
	"fmt"
	"io"
	"strings"
	stdtime "time"
)

// Writer writes a Debian changelog. Entries are written in the order given,
// separated by a blank line; putting them newest first is the caller's job.
//
// The body of an entry is written back exactly as it is held, so a changelog
// read through Reader keeps whatever indentation its generator chose. The
// header and trailer are rewritten in dpkg's layout: the urgency first among
// the header options, and two spaces between the maintainer and an RFC1123
// date with a numeric zone.
type Writer struct {
	w     io.Writer
	wrote bool
	buf   []byte
}

// NewWriter returns a Writer writing an uncompressed changelog to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Write writes one entry. Entries that could not be read back as given are
// rejected rather than written.
func (w *Writer) Write(e Entry) error {
	if err := validate(e); err != nil {
		return err
	}

	// dpkg requires a numeric zone in a changelog date; a zone name is invalid
	// downstream. The date is therefore formatted here rather than through the
	// module's time type, whose MarshalText keeps a zone name when the value
	// carries one - and a date taken from a file's mtime, or from time.Now(),
	// carries the host's local zone, name and all. Values decoded by Reader are
	// already anchored in an unnamed zone and render identically either way.
	date := stdtime.Time(e.Date).Format(stdtime.RFC1123Z)

	w.buf = w.buf[:0]
	if w.wrote {
		w.buf = append(w.buf, '\n')
	}

	w.buf = append(w.buf, e.Source...)
	w.buf = append(w.buf, " ("...)
	w.buf = append(w.buf, e.Version.String()...)
	w.buf = append(w.buf, ')')

	for _, dist := range e.Distributions {
		w.buf = append(w.buf, ' ')
		w.buf = append(w.buf, dist...)
	}

	w.buf = append(w.buf, "; "...)
	w.buf = append(w.buf, options(e)...)
	w.buf = append(w.buf, '\n')

	for _, change := range e.Changes {
		w.buf = append(w.buf, change...)
		w.buf = append(w.buf, '\n')
	}

	w.buf = append(w.buf, trailerPrefix...)
	w.buf = append(w.buf, e.Maintainer...)
	w.buf = append(w.buf, ' ', ' ')
	w.buf = append(w.buf, date...)
	w.buf = append(w.buf, '\n')

	if _, err := w.w.Write(w.buf); err != nil {
		return fmt.Errorf("failed to write changelog entry: %w", err)
	}

	w.wrote = true

	return nil
}

// WriteAll writes a batch of entries.
func (w *Writer) WriteAll(entries []Entry) error {
	for _, e := range entries {
		if err := w.Write(e); err != nil {
			return err
		}
	}

	return nil
}

// options renders the header option list, urgency first. An entry that carries
// no options at all would leave the header ending in a bare semicolon, so it
// falls back to DefaultUrgency.
func options(e Entry) string {
	if e.Urgency == "" && len(e.Options) == 0 {
		return "urgency=" + DefaultUrgency
	}

	var b strings.Builder
	if e.Urgency != "" {
		b.WriteString("urgency=")
		b.WriteString(e.Urgency)
	}

	for _, o := range e.Options {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(o.Key)
		b.WriteByte('=')
		b.WriteString(o.Value)
	}

	return b.String()
}

// validate rejects entries that would not survive a round trip. The framing is
// positional throughout - the header is the only line starting in column 0, the
// trailer the only one starting with a single space and a dash - so anything
// that could be mistaken for either is refused.
func validate(e Entry) error {
	if e.Source == "" || strings.ContainsAny(e.Source, " \t\r\n()") {
		return fmt.Errorf("%w: invalid source %q", ErrInvalidEntry, e.Source)
	}

	if e.Version.Empty() {
		return fmt.Errorf("%w: no version for %q", ErrInvalidEntry, e.Source)
	}

	for _, dist := range e.Distributions {
		if dist == "" || strings.ContainsAny(dist, " \t\r\n;") {
			return fmt.Errorf("%w: invalid distribution %q", ErrInvalidEntry, dist)
		}
	}

	if e.Urgency != "" && strings.ContainsAny(e.Urgency, ",;\r\n") {
		return fmt.Errorf("%w: invalid urgency %q", ErrInvalidEntry, e.Urgency)
	}

	for _, o := range e.Options {
		if o.Key == "" || strings.ContainsAny(o.Key, " \t\r\n,;=") {
			return fmt.Errorf("%w: invalid option key %q", ErrInvalidEntry, o.Key)
		}

		if strings.ContainsAny(o.Value, ",;\r\n") {
			return fmt.Errorf("%w: invalid option value %q", ErrInvalidEntry, o.Value)
		}
	}

	for _, change := range e.Changes {
		if strings.ContainsAny(change, "\r\n") {
			return fmt.Errorf("%w: change line spans lines: %q", ErrInvalidEntry, change)
		}

		if strings.TrimSpace(change) == "" {
			continue
		}

		if change[0] != ' ' && change[0] != '\t' {
			return fmt.Errorf("%w: unindented change line: %q", ErrInvalidEntry, change)
		}

		if strings.HasPrefix(change, trailerPrefix) {
			return fmt.Errorf("%w: change line reads as a trailer: %q", ErrInvalidEntry, change)
		}
	}

	if e.Maintainer == "" || strings.ContainsAny(e.Maintainer, "\r\n") ||
		strings.TrimSpace(e.Maintainer) != e.Maintainer {
		return fmt.Errorf("%w: invalid maintainer %q", ErrInvalidEntry, e.Maintainer)
	}

	if stdtime.Time(e.Date).IsZero() {
		return fmt.Errorf("%w: no date for %q", ErrInvalidEntry, e.Source)
	}

	return nil
}
