// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

// Package contents reads and writes the "Contents" indices of a Debian
// repository (Contents-$arch, Contents-source and their udeb variants).
//
// Unlike the rest of this module the Contents index is not a deb822 document:
// it is a flat two column table, one line per path, with no stanzas, fields or
// continuation lines. It lives here because it is part of the same archive
// layout - the Release file that this module already parses carries the hashes
// of these files.
//
// A line is a path, a run of whitespace, and a comma separated list of
// qualified package names:
//
//	usr/bin/dpkg                                            admin/dpkg
//
// The width and composition of that whitespace run is generator specific and
// carries no meaning: dak pads the path to 55 columns with spaces and follows
// it with a single space, dak's Contents-source writer uses a lone tab, and
// apt-ftparchive pads with tabs. Paths longer than the pad column are followed
// by a single separator, and real archives do contain paths with spaces in
// them, so the columns must be split on the *last* whitespace run of the line,
// never the first.
//
// Neither the reader nor the writer compresses: Contents indices are shipped
// gzip compressed (or lz4, in apt's local cache), and wrapping the stream is
// left to the caller.
package contents

import (
	"errors"
	"strings"
)

// Errors reported by Reader and Writer, wrapped with the offending line.
// Use errors.Is to test for them.
var (
	// ErrInvalidEntry is returned by Reader for a line that does not hold both
	// columns: a path, whitespace, and a non-empty package list.
	ErrInvalidEntry = errors.New("invalid contents entry")

	// ErrInvalidPath is returned by Writer for a path that is empty, carries
	// leading or trailing whitespace, or contains a line break. Such a path
	// cannot be written and read back unambiguously.
	ErrInvalidPath = errors.New("invalid contents path")

	// ErrInvalidPackageList is returned by Writer for an empty package list or
	// for a package name containing a comma, whitespace or a line break.
	ErrInvalidPackageList = errors.New("invalid contents package list")
)

// Entry is one line of a Contents index: an archive relative path and the
// qualified names of the packages shipping it.
type Entry struct {
	// Path is the path the package installs, relative to the filesystem root
	// and without a leading "./" or "/", as it appears in the file.
	Path string

	// Packages holds the qualified package names sharing that path, in the
	// order they appear on the line. Contents-source indices name source
	// packages here, binary indices name binary packages.
	Packages []string
}

// QualifiedName is a package reference as it appears in the second column:
// the package name, prefixed by the section and - depending on the archive -
// the area it sits in. Debian writes "contrib/games/crafty" and "admin/dpkg"
// alike, so neither prefix can be assumed present.
type QualifiedName struct {
	// Area is the archive area ("main", "contrib", "non-free", "universe"),
	// empty when the name carries no area prefix.
	Area string

	// Section is the section the package is filed under ("admin", "games"),
	// empty when the name is a bare package name.
	Section string

	// Name is the package name.
	Name string
}

// ParseQualifiedName splits a qualified package name into its components. The
// name is always the last "/" separated component; anything before the section
// is taken as the area, so that an unexpectedly deep name still yields the
// package name rather than a truncated one.
func ParseQualifiedName(s string) QualifiedName {
	var q QualifiedName

	rest, name, ok := cutLast(s, "/")
	q.Name = name
	if !ok {
		q.Name = s

		return q
	}

	area, section, ok := cutLast(rest, "/")
	if !ok {
		q.Section = rest

		return q
	}

	q.Area, q.Section = area, section

	return q
}

// String renders the qualified name back into its "[[area/]section/]name"
// wire form.
func (q QualifiedName) String() string {
	switch {
	case q.Area != "" && q.Section != "":
		return q.Area + "/" + q.Section + "/" + q.Name
	case q.Section != "":
		return q.Section + "/" + q.Name
	default:
		return q.Name
	}
}

// cutLast slices s around the last instance of sep.
func cutLast(s, sep string) (before, after string, found bool) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, "", false
	}

	return s[:i], s[i+len(sep):], true
}
