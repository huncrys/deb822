// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

// Package changelog reads and writes Debian changelogs: debian/changelog in a
// source package, and the changelog.Debian.gz that a binary package ships in
// /usr/share/doc.
//
// Like the Contents indices this is not a deb822 document - it has no stanzas,
// fields or continuation lines - but it is the file every deb822 document in
// this module is ultimately derived from. dpkg-genchanges lifts the version,
// distribution, urgency, date and change text of a .changes file straight out
// of the topmost entry, and dpkg-source does the same for a .dsc.
//
// An entry is a header line, a body, and a trailer line:
//
//	hello (2.10-3) unstable; urgency=medium
//
//	  * Add some autopkgtests. Closes: #871622.
//
//	 -- Santiago Vila <sanvila@debian.org>  Mon, 26 Dec 2022 16:30:00 +0100
//
// Entries are separated by a single blank line and run newest first. The header
// line starts in column 0, every body line starts with whitespace, and the
// trailer starts with exactly one space followed by "--".
//
// # Generators disagree about the body
//
// The layout inside an entry is not fixed. dpkg and dch put a blank line
// between the header and the body; goreleaser/chglog, which nFPM uses to build
// a changelog from git history, does not, and indents its continuation lines by
// three spaces rather than the two dpkg emits. Real files also carry trailing
// whitespace and markdown. None of that is meaningful, and none of it may be
// normalised away, so Entry.Changes holds the body exactly as it was read -
// blank lines, indentation and all - and the Writer plays it back untouched.
//
// # Dates are re-emitted canonically
//
// The trailer date is decoded through the module's [time.Time], which accepts
// the layouts that turn up in the archive - space padded and unpadded days
// among them - and re-emits an RFC1123 date with a numeric zone, which is what
// dpkg requires. A changelog written before that convention settled therefore
// does not round trip byte for byte; reformatting it once reaches a fixed point
// that does.
//
// Neither the reader nor the writer compresses: binary packages ship the
// changelog gzipped, and wrapping the stream is left to the caller.
package changelog

import (
	"errors"
	"strings"

	deb822time "oaklab.hu/debian/deb822/types/time"
	"oaklab.hu/debian/deb822/types/version"
)

// DefaultUrgency is the urgency dch assigns to a new entry, and the one the
// Writer falls back to for an entry that carries no header options at all.
const DefaultUrgency = "medium"

// Errors reported by Reader and Writer, wrapped with the offending line.
// Use errors.Is to test for them.
var (
	// ErrNotDebianFormat is returned by Reader when the first non-blank line of
	// the stream is not an entry header. Binary packages are free to ship an
	// upstream changelog in whatever shape upstream keeps it, so a consumer
	// walking an archive should expect this and fall back to treating the file
	// as opaque text rather than as a failure.
	ErrNotDebianFormat = errors.New("not a Debian changelog")

	// ErrInvalidHeader is returned by Reader for a line that has the shape of an
	// entry header but cannot be read as one, most often because the version in
	// parentheses is not a valid dpkg version.
	ErrInvalidHeader = errors.New("invalid changelog header")

	// ErrInvalidTrailer is returned by Reader for a trailer line that does not
	// yield both a maintainer and a date.
	ErrInvalidTrailer = errors.New("invalid changelog trailer")

	// ErrMissingTrailer is returned by Reader when a new entry begins before the
	// current one has been closed by a trailer.
	ErrMissingTrailer = errors.New("changelog entry has no trailer")

	// ErrInvalidEntry is returned by Writer for an entry that could not be read
	// back as it was given.
	ErrInvalidEntry = errors.New("invalid changelog entry")
)

// Entry is one changelog entry.
type Entry struct {
	// Source is the name of the source package the entry belongs to. It is
	// allowed to change over the life of a package, and does in real files.
	Source string

	// Version is the version this entry released.
	Version version.Version

	// Distributions holds the suites the upload was targeted at, in the order
	// written - usually the single "unstable", or "UNRELEASED" while the entry
	// is still being worked on. Ancient entries omit it entirely.
	Distributions []string

	// Urgency is the value of the "urgency" header option ("low", "medium",
	// "high", "emergency", "critical"), empty when the entry carries none. The
	// Writer always emits it first, which is where every generator puts it.
	Urgency string

	// Options holds the remaining header options in the order written, such as
	// the "binary-only=yes" that marks a binNMU changelog.
	Options []Option

	// Changes holds the body lines verbatim, exactly as they sit between the
	// header and the trailer, including the blank lines that delimit them.
	Changes []string

	// Maintainer is the name and email address from the trailer, in
	// "Name <email>" form. It is who uploaded this version, which for an NMU or
	// a binNMU is not the package's maintainer.
	Maintainer string

	// Date is the trailer date.
	Date deb822time.Time
}

// Option is one "key=value" pair of the header line's option list.
type Option struct {
	Key   string
	Value string
}

// Option returns the value of the named header option. Urgency is reported
// through this too, so that a caller walking the options does not have to
// special-case it.
func (e Entry) Option(key string) (string, bool) {
	if strings.EqualFold(key, "urgency") {
		return e.Urgency, e.Urgency != ""
	}

	for _, o := range e.Options {
		if strings.EqualFold(o.Key, key) {
			return o.Value, true
		}
	}

	return "", false
}

// BinaryOnly reports whether the entry is the "binary-only=yes" pseudo entry
// that dpkg prepends for a binNMU. Such an entry does not exist in the source
// package, and its version carries the +bN suffix.
func (e Entry) BinaryOnly() bool {
	v, ok := e.Option("binary-only")

	return ok && strings.EqualFold(v, "yes")
}

// Text returns the body as a single string with the delimiting blank lines
// trimmed off, which is the form dpkg-genchanges folds into the Changes field
// of a .changes file.
func (e Entry) Text() string {
	lines := e.Changes
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}
