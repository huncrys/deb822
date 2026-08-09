// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package changelog_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	stdtime "time"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822/changelog"
	deb822time "oaklab.hu/debian/deb822/types/time"
	"oaklab.hu/debian/deb822/types/version"
)

// TestReadHello reads hello's changelog as the Debian archive ships it. The
// package has been maintained since 1995, so the file carries most of the shape
// the format has drifted through.
func TestReadHello(t *testing.T) {
	entries, r := readFile(t, "../testdata/hello_2.10-3_changelog")

	require.Len(t, entries, 39)

	// The current entry, in the layout dpkg and dch write: a blank line after
	// the header, the changes indented by two spaces, a blank line before the
	// trailer.
	require.Equal(t, "hello", entries[0].Source)
	require.Equal(t, version.Version{Version: "2.10", Revision: "3"}, entries[0].Version)
	require.Equal(t, []string{"unstable"}, entries[0].Distributions)
	require.Equal(t, "medium", entries[0].Urgency)
	require.Equal(t, "Santiago Vila <sanvila@debian.org>", entries[0].Maintainer)
	require.Equal(t,
		stdtime.Date(2022, stdtime.December, 26, 16, 30, 0, 0, stdtime.FixedZone("", 3600)),
		stdtime.Time(entries[0].Date))
	require.Equal(t, "", entries[0].Changes[0])
	require.Equal(t, "  * Add some autopkgtests. Closes: #871622.", entries[0].Changes[1])
	require.Equal(t, "", entries[0].Changes[len(entries[0].Changes)-1])

	// The source package was renamed twice over its life, so the name on the
	// header line is not a constant.
	require.Equal(t, "hello-debhelper", entries[5].Source)

	// An upload targeted at two suites at once.
	require.Equal(t, []string{"frozen", "unstable"}, entries[23].Distributions)

	// Urgency was not always spelled in lowercase.
	require.Equal(t, "LOW", entries[34].Urgency)

	// The oldest entries name no distribution at all and carry a "priority"
	// option in place of the urgency dpkg later settled on.
	require.Empty(t, entries[37].Distributions)
	require.Equal(t, "", entries[37].Urgency)
	require.Equal(t, []changelog.Option{{Key: "priority", Value: "LOW"}}, entries[37].Options)

	// A pre-format history sits past the last entry. It is kept aside rather
	// than parsed, and rather than failing the read.
	trailing := r.Trailing()
	require.Len(t, trailing, 16)
	require.Equal(t, "Hello 1.3 Debian 3 - iwj", trailing[0])
	require.Equal(t, "Initial release.", trailing[len(trailing)-1])
}

// TestReadChglog reads a slice of an atuin-client changelog built by nFPM out
// of goreleaser/chglog, which lays an entry out differently to dpkg: no blank
// line after the header, and continuations indented by three spaces.
func TestReadChglog(t *testing.T) {
	entries, r := readFile(t, "../testdata/atuin-client_18.19.0.changelog")

	require.Len(t, entries, 2)
	require.Empty(t, r.Trailing())

	require.Equal(t, "atuin-client", entries[0].Source)
	require.Equal(t, version.Version{Version: "18.19.0"}, entries[0].Version)
	require.Equal(t, "low", entries[0].Urgency)
	require.Equal(t, "Ellie Huxtable <ellie@atuin.sh>", entries[0].Maintainer)

	// The body opens on the line straight after the header.
	require.Equal(t, "  * chore(release): prepare for release 18.19.0 (#3847)", entries[0].Changes[0])
	require.Equal(t, "   - ### Bug Fixes", entries[0].Changes[1])

	require.Equal(t, "atuin-bot <152089506+atuin-bot@users.noreply.github.com>", entries[1].Maintainer)
}

// TestWriteIsByteStable pins that decoding and re-encoding a changelog written
// by current tooling reproduces it byte for byte, layout quirks and all.
func TestWriteIsByteStable(t *testing.T) {
	original, err := os.ReadFile("../testdata/atuin-client_18.19.0.changelog")
	require.NoError(t, err)

	entries, err := changelog.NewReader(bytes.NewReader(original)).ReadAll()
	require.NoError(t, err)

	require.Equal(t, string(original), write(t, entries))
}

// TestReformatIsIdempotent pins the fixed point for a file that predates the
// current conventions. hello's oldest trailers carry space padded and unpadded
// days, which are re-emitted in the canonical layout, so the first write is not
// byte identical - but every write after it is.
func TestReformatIsIdempotent(t *testing.T) {
	original, err := os.ReadFile("../testdata/hello_2.10-3_changelog")
	require.NoError(t, err)

	entries, err := changelog.NewReader(bytes.NewReader(original)).ReadAll()
	require.NoError(t, err)

	first := write(t, entries)
	require.NotEqual(t, string(original), first)
	require.Contains(t, string(original), "  Sat,  9 Dec 2006 17:00:14 +0100")
	require.Contains(t, first, "  Sat, 09 Dec 2006 17:00:14 +0100")

	reread, err := changelog.NewReader(strings.NewReader(first)).ReadAll()
	require.NoError(t, err)

	require.Equal(t, first, write(t, reread))
}

// TestReadUpstreamFormat covers the changelog.gz that a binary package ships
// when it has no changelog.Debian: upstream's own file, in whatever shape
// upstream keeps it. Telling that apart from a corrupt Debian changelog is what
// lets a consumer publish the file unparsed instead of rejecting the package.
func TestReadUpstreamFormat(t *testing.T) {
	const upstream = `0.41.1
---
* **machine**
  - esp32c3: correct pin interrupt setup call that was overlooked from #5320
`

	_, err := changelog.NewReader(strings.NewReader(upstream)).ReadAll()
	require.ErrorIs(t, err, changelog.ErrNotDebianFormat)
}

// TestReadSalvagedDates covers the trailer dates the archive still carries
// because dpkg-parsechangelog never parses that field. Both of these are real
// lines out of bash's changelog.
func TestReadSalvagedDates(t *testing.T) {
	for _, tc := range []struct {
		trailer string
		want    stdtime.Time
	}{
		{
			// A full month name.
			trailer: " -- Joel Klecker <jk@espy.org>  Tue, 14 July 1998 16:26:43 -0700",
			want: stdtime.Date(1998, stdtime.July, 14, 16, 26, 43, 0,
				stdtime.FixedZone("", -7*3600)),
		},
		{
			// A four letter weekday, and the wrong one at that - 19 June 1997
			// was a Thursday, but 1997 is not what "Thur" abbreviates.
			trailer: " -- James Troup <jjtroup@comp.brad.ac.uk>  Thur, 19 June 1997 19:13:34 +0100",
			want: stdtime.Date(1997, stdtime.June, 19, 19, 13, 34, 0,
				stdtime.FixedZone("", 3600)),
		},
	} {
		t.Run(tc.trailer, func(t *testing.T) {
			entries, err := changelog.NewReader(strings.NewReader(
				"bash (1.0) unstable; urgency=low\n\n  * Something.\n\n" + tc.trailer + "\n",
			)).ReadAll()
			require.NoError(t, err)
			require.Len(t, entries, 1)

			require.Equal(t, tc.want, stdtime.Time(entries[0].Date))
		})
	}
}

func TestReadErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		err  error
	}{
		{
			name: "version that dpkg would reject",
			in:   "hello (1.0-) unstable; urgency=low\n\n  * Nope.\n\n -- A B <a@b.c>  Mon, 26 Dec 2022 16:30:00 +0100\n",
			err:  changelog.ErrInvalidHeader,
		},
		{
			name: "trailer without a date",
			in:   "hello (1.0) unstable; urgency=low\n\n  * Nope.\n\n -- A B <a@b.c>\n",
			err:  changelog.ErrInvalidTrailer,
		},
		{
			name: "trailer date in no known layout",
			in:   "hello (1.0) unstable; urgency=low\n\n  * Nope.\n\n -- A B <a@b.c>  2022-12-26\n",
			err:  changelog.ErrInvalidTrailer,
		},
		{
			name: "entry closed by the next entry",
			in:   "hello (1.0) unstable; urgency=low\n\n  * Nope.\n\nhello (0.9) unstable; urgency=low\n\n  * Nope.\n\n -- A B <a@b.c>  Mon, 26 Dec 2022 16:30:00 +0100\n",
			err:  changelog.ErrMissingTrailer,
		},
		{
			name: "entry running into the end of the file",
			in:   "hello (1.0) unstable; urgency=low\n\n  * Nope.\n",
			err:  changelog.ErrMissingTrailer,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := changelog.NewReader(strings.NewReader(tc.in)).ReadAll()
			require.ErrorIs(t, err, tc.err)
		})
	}
}

// TestReadBinaryOnly covers the pseudo entry dpkg prepends for a binNMU, where
// the urgency is followed by a second header option.
func TestReadBinaryOnly(t *testing.T) {
	const binNMU = `mutt (2.2.12-0.1+b1) sid; urgency=low, binary-only=yes

  * Binary-only non-maintainer upload for amd64; no source changes.
  * Rebuild against libgnutls30t64.

 -- amd64 Build Daemon (x86-grnet-01) <buildd_amd64-x86-grnet-01@buildd.debian.org>  Sat, 20 Apr 2024 10:11:04 +0000

mutt (2.2.12-0.1) unstable; urgency=medium

  * Non-maintainer upload.

 -- Sebastian Ramacher <sramacher@debian.org>  Sun, 07 Apr 2024 22:15:39 +0200
`

	entries, err := changelog.NewReader(strings.NewReader(binNMU)).ReadAll()
	require.NoError(t, err)
	require.Len(t, entries, 2)

	require.True(t, entries[0].BinaryOnly())
	require.Equal(t, "low", entries[0].Urgency)
	require.Equal(t, []changelog.Option{{Key: "binary-only", Value: "yes"}}, entries[0].Options)

	urgency, ok := entries[0].Option("Urgency")
	require.True(t, ok)
	require.Equal(t, "low", urgency)

	require.False(t, entries[1].BinaryOnly())

	// The header option order is part of the wire form and is preserved.
	require.Equal(t, binNMU, write(t, entries))
}

// TestWritePreservesBody pins that the body is played back untouched, down to
// the trailing whitespace and unconventional indentation that real generators
// leave behind.
func TestWritePreservesBody(t *testing.T) {
	const ragged = "btop (1.4.7) unstable; urgency=low\n" +
		"  * `cpu responsive` was renamed to `cpu direct` elsewhere, \n" +
		"\t* tab indented, because nothing forbids it\n" +
		"\n" +
		"   - three spaces, chglog style\n" +
		"\n" +
		" -- aristocratos <gnmjpl@gmail.com>  Fri, 01 May 2026 16:05:43 +0000\n"

	entries, err := changelog.NewReader(strings.NewReader(ragged)).ReadAll()
	require.NoError(t, err)
	require.Len(t, entries, 1)

	require.Equal(t, ragged, write(t, entries))

	// Text drops the delimiting blank lines but nothing between them, which is
	// the body as dpkg-genchanges folds it into a .changes file.
	require.Equal(t,
		"  * `cpu responsive` was renamed to `cpu direct` elsewhere, \n"+
			"\t* tab indented, because nothing forbids it\n"+
			"\n"+
			"   - three spaces, chglog style",
		entries[0].Text())
}

// TestWriteSynthesised covers building an entry from scratch, the way a
// repository builder does when a package ships no changelog of its own.
func TestWriteSynthesised(t *testing.T) {
	var buf bytes.Buffer

	require.NoError(t, changelog.NewWriter(&buf).Write(changelog.Entry{
		Source:        "composer-phar",
		Version:       version.MustParse("2.9.8"),
		Distributions: []string{"unstable"},
		Urgency:       "medium",
		Changes:       []string{"", "  * No changelog available.", ""},
		Maintainer:    "Kristof Bach <crys@crys.hu>",
		Date: deb822time.Time(stdtime.Date(2026, stdtime.May, 13, 18, 4, 5, 0,
			stdtime.FixedZone("", 2*3600))),
	}))

	require.Equal(t, "composer-phar (2.9.8) unstable; urgency=medium\n"+
		"\n"+
		"  * No changelog available.\n"+
		"\n"+
		" -- Kristof Bach <crys@crys.hu>  Wed, 13 May 2026 18:04:05 +0200\n", buf.String())
}

// TestWriteDefaultsUrgency covers the one case where the writer supplies
// something the caller did not: an entry with no header options at all would
// otherwise end in a bare semicolon.
func TestWriteDefaultsUrgency(t *testing.T) {
	var buf bytes.Buffer

	require.NoError(t, changelog.NewWriter(&buf).Write(changelog.Entry{
		Source:        "hello",
		Version:       version.MustParse("1.0"),
		Distributions: []string{"unstable"},
		Maintainer:    "A B <a@b.c>",
		Date:          deb822time.Time(stdtime.Date(2026, stdtime.May, 13, 18, 4, 5, 0, stdtime.UTC)),
	}))

	require.Contains(t, buf.String(), "hello (1.0) unstable; urgency="+changelog.DefaultUrgency+"\n")
}

func TestWriteErrors(t *testing.T) {
	valid := changelog.Entry{
		Source:        "hello",
		Version:       version.MustParse("1.0"),
		Distributions: []string{"unstable"},
		Urgency:       "medium",
		Changes:       []string{"", "  * Something.", ""},
		Maintainer:    "A B <a@b.c>",
		Date:          deb822time.Time(stdtime.Date(2026, stdtime.May, 13, 18, 4, 5, 0, stdtime.UTC)),
	}

	for _, tc := range []struct {
		name  string
		mutet func(*changelog.Entry)
	}{
		{"no source", func(e *changelog.Entry) { e.Source = "" }},
		{"source with a space", func(e *changelog.Entry) { e.Source = "hello world" }},
		{"no version", func(e *changelog.Entry) { e.Version = version.Version{} }},
		{"distribution with a semicolon", func(e *changelog.Entry) { e.Distributions = []string{"a;b"} }},
		{"urgency with a comma", func(e *changelog.Entry) { e.Urgency = "medium, binary-only=yes" }},
		{"option key with an equals sign", func(e *changelog.Entry) {
			e.Options = []changelog.Option{{Key: "a=b", Value: "c"}}
		}},
		{"unindented change line", func(e *changelog.Entry) {
			e.Changes = []string{"hello (2.0) unstable; urgency=low"}
		}},
		{"change line reading as a trailer", func(e *changelog.Entry) {
			e.Changes = []string{" -- A B <a@b.c>  Wed, 13 May 2026 18:04:05 +0000"}
		}},
		{"change line spanning lines", func(e *changelog.Entry) { e.Changes = []string{"  * a\n  * b"} }},
		{"no maintainer", func(e *changelog.Entry) { e.Maintainer = "" }},
		{"no date", func(e *changelog.Entry) { e.Date = deb822time.Time{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry := valid
			tc.mutet(&entry)

			require.ErrorIs(t, changelog.NewWriter(io.Discard).Write(entry), changelog.ErrInvalidEntry)
		})
	}
}

// TestRejectedEntryIsNotWritten pins that a rejected entry leaves the stream
// untouched, so that the blank line separating entries is not emitted for it.
func TestRejectedEntryIsNotWritten(t *testing.T) {
	var buf bytes.Buffer

	w := changelog.NewWriter(&buf)
	require.ErrorIs(t, w.Write(changelog.Entry{}), changelog.ErrInvalidEntry)
	require.Empty(t, buf.String())
}

func readFile(t *testing.T, path string) ([]changelog.Entry, *changelog.Reader) {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, f.Close()) })

	r := changelog.NewReader(f)

	entries, err := r.ReadAll()
	require.NoError(t, err)

	return entries, r
}

func write(t *testing.T, entries []changelog.Entry) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, changelog.NewWriter(&buf).WriteAll(entries))

	return buf.String()
}
