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
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	deb822time "oaklab.hu/debian/deb822/types/time"
	"oaklab.hu/debian/deb822/types/version"
)

// maxLineLength bounds a single line. Change lines are prose and wrap well
// below this; a megabyte keeps a corrupt or binary stream from being buffered
// whole.
const maxLineLength = 1 << 20

// trailerPrefix opens a trailer line. The single leading space is what tells it
// apart from a change line, which every generator indents by at least two.
const trailerPrefix = " -- "

// line is a scanned line together with its 1-based number, kept so that errors
// can name the offending line.
type line struct {
	no   int
	text string
}

// Reader reads a Debian changelog one entry at a time, newest first.
type Reader struct {
	s        *bufio.Scanner
	no       int
	started  bool
	trailing []string
}

// NewReader returns a Reader reading the uncompressed changelog in r.
func NewReader(r io.Reader) *Reader {
	s := bufio.NewScanner(r)
	s.Buffer(nil, maxLineLength)

	return &Reader{s: s}
}

// Read returns the next entry. It returns io.EOF when the changelog is
// exhausted, and ErrNotDebianFormat if the stream never was one.
func (r *Reader) Read() (Entry, error) {
	for {
		l, ok := r.scan()
		if !ok {
			return Entry{}, r.eof()
		}

		if strings.TrimSpace(l.text) == "" {
			continue
		}

		if !looksLikeHeader(l.text) {
			// Old changelogs trail off into the free-form history that predates
			// the format, and some carry an editor's local-variables block. Both
			// sit past the last entry, so everything from here on is kept aside
			// rather than parsed - but only once an entry has been read, since a
			// file that starts this way is not a Debian changelog at all.
			if !r.started {
				return Entry{}, fmt.Errorf("line %d: %w: %q", l.no, ErrNotDebianFormat, l.text)
			}

			r.trailing = append(r.trailing, l.text)
			for {
				t, ok := r.scan()
				if !ok {
					return Entry{}, r.eof()
				}

				r.trailing = append(r.trailing, t.text)
			}
		}

		entry, err := parseHeader(l.text)
		if err != nil {
			return Entry{}, fmt.Errorf("line %d: %w: %q", l.no, err, l.text)
		}

		if err := r.readBody(&entry); err != nil {
			return Entry{}, err
		}

		r.started = true

		return entry, nil
	}
}

// ReadAll reads the remaining entries of the changelog.
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

// Trailing returns the lines past the last entry that are not part of the
// changelog proper, verbatim. It is empty for the well-formed files current
// tooling writes, and is only populated once reading has reached the end.
func (r *Reader) Trailing() []string {
	return r.trailing
}

// readBody consumes the change lines and the trailer that close an entry.
func (r *Reader) readBody(entry *Entry) error {
	for {
		l, ok := r.scan()
		if !ok {
			if err := r.s.Err(); err != nil {
				return fmt.Errorf("failed to read changelog: %w", err)
			}

			return fmt.Errorf("line %d: %w: %q", r.no, ErrMissingTrailer, entry.Source)
		}

		if strings.HasPrefix(l.text, trailerPrefix) {
			maintainer, date, err := parseTrailer(l.text)
			if err != nil {
				return fmt.Errorf("line %d: %w: %q", l.no, err, l.text)
			}

			entry.Maintainer, entry.Date = maintainer, date

			return nil
		}

		if looksLikeHeader(l.text) {
			return fmt.Errorf("line %d: %w: %q", l.no, ErrMissingTrailer, entry.Source)
		}

		entry.Changes = append(entry.Changes, l.text)
	}
}

// eof reports the end of the stream, or the scanner's error if it stopped for
// another reason.
func (r *Reader) eof() error {
	if err := r.s.Err(); err != nil {
		return fmt.Errorf("failed to read changelog: %w", err)
	}

	return io.EOF
}

// scan pulls one line off the underlying scanner.
func (r *Reader) scan() (line, bool) {
	if !r.s.Scan() {
		return line{}, false
	}
	r.no++

	return line{no: r.no, text: r.s.Text()}, true
}

// looksLikeHeader reports whether a line has the shape of an entry header:
// a name in column 0, a parenthesised version, and a semicolon closing the
// distribution list. It only frames the entry - parseHeader decides whether the
// pieces are usable - so that garbage is told apart from a malformed entry.
func looksLikeHeader(text string) bool {
	_, _, _, ok := splitHeader(text)

	return ok
}

// splitHeader cuts a header line into its three parts without validating them.
func splitHeader(text string) (source, ver, rest string, ok bool) {
	if text == "" || text[0] == ' ' || text[0] == '\t' {
		return "", "", "", false
	}

	source, after, ok := strings.Cut(text, "(")
	if !ok {
		return "", "", "", false
	}

	ver, rest, ok = strings.Cut(after, ")")
	if !ok {
		return "", "", "", false
	}

	source = strings.TrimSpace(source)
	if source == "" || strings.ContainsAny(source, " \t") {
		return "", "", "", false
	}

	if ver = strings.TrimSpace(ver); ver == "" || strings.ContainsAny(ver, " \t(") {
		return "", "", "", false
	}

	// The semicolon is mandatory: without it there is nothing to separate a
	// header from a stray line that happens to hold a bracketed word.
	if !strings.Contains(rest, ";") {
		return "", "", "", false
	}

	return source, ver, rest, true
}

// parseHeader reads a header line into an entry. The distribution list is
// allowed to be empty: entries from the mid-nineties omit it, writing only
// "hello (1.3-6); priority=LOW".
func parseHeader(text string) (Entry, error) {
	source, ver, rest, ok := splitHeader(text)
	if !ok {
		return Entry{}, ErrInvalidHeader
	}

	parsed, err := version.Parse(ver)
	if err != nil {
		return Entry{}, fmt.Errorf("%w: %w", ErrInvalidHeader, err)
	}

	dists, opts, _ := strings.Cut(rest, ";")

	entry := Entry{
		Source:        source,
		Version:       parsed,
		Distributions: strings.Fields(dists),
	}

	for _, o := range parseOptions(opts) {
		if strings.EqualFold(o.Key, "urgency") {
			entry.Urgency = o.Value

			continue
		}

		entry.Options = append(entry.Options, o)
	}

	return entry, nil
}

// parseOptions splits the comma separated "key=value" list that closes a header
// line. A bare word is kept as a key with an empty value rather than dropped.
func parseOptions(text string) []Option {
	var opts []Option
	for field := range strings.SplitSeq(text, ",") {
		if strings.TrimSpace(field) == "" {
			continue
		}

		key, value, _ := strings.Cut(field, "=")
		if key = strings.TrimSpace(key); key == "" {
			continue
		}

		opts = append(opts, Option{Key: key, Value: strings.TrimSpace(value)})
	}

	return opts
}

// parseTrailer reads the maintainer and date off a trailer line. The two are
// separated by two spaces in everything dpkg writes, but one space also occurs,
// so the split is made at the end of the email address instead.
func parseTrailer(text string) (string, deb822time.Time, error) {
	rest := text[len(trailerPrefix):]

	var maintainer, date string
	if i := strings.LastIndex(rest, ">"); i >= 0 {
		maintainer, date = rest[:i+1], rest[i+1:]
	} else if i := strings.LastIndex(rest, "  "); i >= 0 {
		maintainer, date = rest[:i], rest[i:]
	} else {
		return "", deb822time.Time{}, ErrInvalidTrailer
	}

	maintainer, date = strings.TrimSpace(maintainer), strings.TrimSpace(date)
	if maintainer == "" || date == "" {
		return "", deb822time.Time{}, ErrInvalidTrailer
	}

	var parsed deb822time.Time

	err := parsed.UnmarshalText([]byte(date))
	if err != nil {
		if salvaged, ok := salvageDate(date); ok {
			err = parsed.UnmarshalText([]byte(salvaged))
		}
	}

	if err != nil {
		return "", deb822time.Time{}, fmt.Errorf("%w: %w", ErrInvalidTrailer, err)
	}

	return maintainer, parsed, nil
}

// salvageDate rewrites the decorative part of a trailer date that no layout
// accepts. dpkg-parsechangelog never parses this field - it hands the string
// through to the .changes file untouched - so the archive carries dates no date
// library will take: bash still ships "Thur, 19 June 1997 19:13:34 +0100", with
// a four letter weekday and a full month name. The weekday is redundant with
// the date and is dropped; the month is cut to the three letter form, which for
// every English month name is its standard abbreviation.
func salvageDate(date string) (string, bool) {
	fields := strings.Fields(date)
	if len(fields) < 4 {
		return "", false
	}

	if strings.HasSuffix(fields[0], ",") {
		fields = fields[1:]
		if len(fields) < 4 {
			return "", false
		}
	}

	if len(fields[1]) > 3 {
		fields[1] = fields[1][:3]
	}

	return strings.Join(fields, " "), true
}
