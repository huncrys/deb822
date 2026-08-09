// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package contents_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822/contents"
)

// TestReadDak reads a slice of Debian bookworm's contrib/Contents-all, as
// written by dak: the path padded to 55 columns with spaces.
func TestReadDak(t *testing.T) {
	entries := readFile(t, "../testdata/Contents-all")

	require.Len(t, entries, 85)

	require.Equal(t, contents.Entry{
		Path:     "etc/apache2/conf-available/mathjax-siunitx.conf",
		Packages: []string{"contrib/javascript/mathjax-siunitx"},
	}, entries[0])

	// A path shared by several packages.
	require.Contains(t, entries, contents.Entry{
		Path: "usr/share/crafty/book.bin",
		Packages: []string{
			"contrib/games/crafty-books-medium",
			"contrib/games/crafty-books-medtosmall",
			"contrib/games/crafty-books-small",
		},
	})

	// A path containing a space, still padded out to the separator column.
	require.Contains(t, entries, contents.Entry{
		Path:     "usr/share/games/corsix-th/Campaigns/Four Corners.map",
		Packages: []string{"contrib/games/corsix-th-data"},
	})

	// A path containing a space that is itself 55 columns wide, so that the
	// separator degenerates to the single space that also sits in the path.
	// Splitting on the first whitespace run would mangle this line.
	require.Contains(t, entries, contents.Entry{
		Path:     "usr/share/games/corsix-th/Campaigns/Greyham Gardens.map",
		Packages: []string{"contrib/games/corsix-th-data"},
	})

	// A path longer than the pad column, separated by a single space.
	require.Contains(t, entries, contents.Entry{
		Path: "usr/share/festival/voices/english/rab_diphone/festvox/rab_diphone.scm",
		Packages: []string{
			"contrib/sound/festvox-rablpc16k",
			"contrib/sound/festvox-rablpc8k",
		},
	})
}

// TestReadAptFtparchive reads a slice of a Contents-amd64 written by
// apt-ftparchive, which pads with tabs rather than spaces.
func TestReadAptFtparchive(t *testing.T) {
	entries := readFile(t, "../testdata/Contents-amd64")

	require.Len(t, entries, 40)

	require.Equal(t, contents.Entry{
		Path:     "usr/libexec/docker/cli-plugins/docker-compose",
		Packages: []string{"admin/docker-compose-plugin"},
	}, entries[0])

	for _, e := range entries {
		require.NotContains(t, e.Path, "\t", "padding leaked into the path")
		for _, name := range e.Packages {
			require.NotContains(t, name, "\t", "padding leaked into the package list")
		}
	}
}

func TestReadLegacyHeader(t *testing.T) {
	f, err := os.Open("../testdata/Contents-legacy-header")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	r := contents.NewReader(f)

	entries, err := r.ReadAll()
	require.NoError(t, err)

	require.Equal(t, []contents.Entry{
		{Path: "bin/bash", Packages: []string{"shells/bash"}},
		{Path: "usr/bin/dpkg", Packages: []string{"admin/dpkg"}},
		{Path: "usr/share/doc/README.Debian", Packages: []string{"admin/dpkg", "admin/dpkg-dev"}},
	}, entries)

	header := r.Header()
	require.Len(t, header, 12)
	require.Equal(t, "This file maps each file available in the Debian GNU/Linux system to", header[0])
	require.NotContains(t, strings.Join(header, "\n"), "LOCATION")
}

// TestReadNoHeader pins that a headerless file is not eaten by the header
// lookahead, including one longer than the lookahead window.
func TestReadNoHeader(t *testing.T) {
	var sb strings.Builder
	for i := range 250 {
		require.NoError(t, contents.NewWriter(&sb).Write(contents.Entry{
			Path:     "usr/share/doc/pkg/file" + strings.Repeat("x", i%7),
			Packages: []string{"admin/pkg"},
		}))
	}

	r := contents.NewReader(strings.NewReader(sb.String()))

	entries, err := r.ReadAll()
	require.NoError(t, err)
	require.Len(t, entries, 250)
	require.Empty(t, r.Header())
}

func TestReadSkipsBlankLines(t *testing.T) {
	entries, err := contents.NewReader(strings.NewReader(
		"\nusr/bin/dpkg admin/dpkg\r\n\n   \nbin/bash shells/bash\n",
	)).ReadAll()
	require.NoError(t, err)

	require.Equal(t, []contents.Entry{
		{Path: "usr/bin/dpkg", Packages: []string{"admin/dpkg"}},
		{Path: "bin/bash", Packages: []string{"shells/bash"}},
	}, entries)
}

func TestReadInvalidEntry(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
	}{
		{"no separator", "usr/bin/dpkg admin/dpkg\nusr/bin/nonsense\n"},
		{"no package list", "usr/bin/dpkg admin/dpkg\nusr/bin/nonsense \n"},
		{"empty package list", "usr/bin/dpkg admin/dpkg\nusr/bin/nonsense ,,\n"},
		{"no path", "usr/bin/dpkg admin/dpkg\n admin/dpkg\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := contents.NewReader(strings.NewReader(tc.in)).ReadAll()
			require.ErrorIs(t, err, contents.ErrInvalidEntry)
			require.Contains(t, err.Error(), "line 2:")
		})
	}
}

// TestWriteIsByteStable pins that decoding and re-encoding a real dak written
// index reproduces it byte for byte.
func TestWriteIsByteStable(t *testing.T) {
	for _, path := range []string{"../testdata/Contents-all", "../testdata/Contents-amd64"} {
		t.Run(path, func(t *testing.T) {
			original, err := os.ReadFile(path)
			require.NoError(t, err)

			entries := readFile(t, path)

			var buf bytes.Buffer
			require.NoError(t, contents.NewWriter(&buf).WriteAll(entries))

			if strings.Contains(string(original), "\t") {
				// apt-ftparchive pads with tabs, so only the entries survive a
				// round trip, not the whitespace. Re-reading is the fixed point.
				reread, err := contents.NewReader(&buf).ReadAll()
				require.NoError(t, err)
				require.Equal(t, entries, reread)

				return
			}

			require.Equal(t, string(original), buf.String())
		})
	}
}

func TestWriteLayouts(t *testing.T) {
	entry := contents.Entry{
		Path:     "usr/bin/dpkg",
		Packages: []string{"admin/dpkg", "admin/dpkg-dev"},
	}

	for _, tc := range []struct {
		name string
		opts []contents.WriterOption
		want string
	}{
		{
			name: "dak binary",
			want: "usr/bin/dpkg                                            admin/dpkg,admin/dpkg-dev\n",
		},
		{
			name: "dak source",
			opts: []contents.WriterOption{contents.WithPadding(0), contents.WithTabSeparator()},
			want: "usr/bin/dpkg\tadmin/dpkg,admin/dpkg-dev\n",
		},
		{
			name: "unpadded",
			opts: []contents.WriterOption{contents.WithPadding(0)},
			want: "usr/bin/dpkg admin/dpkg,admin/dpkg-dev\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, contents.NewWriter(&buf, tc.opts...).Write(entry))
			require.Equal(t, tc.want, buf.String())

			entries, err := contents.NewReader(strings.NewReader(buf.String())).ReadAll()
			require.NoError(t, err)
			require.Equal(t, []contents.Entry{entry}, entries)
		})
	}
}

func TestWriteRejectsAmbiguousEntries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		entry contents.Entry
		want  error
	}{
		{"empty path", contents.Entry{Path: "", Packages: []string{"admin/dpkg"}}, contents.ErrInvalidPath},
		{"trailing space", contents.Entry{Path: "usr/bin/dpkg ", Packages: []string{"admin/dpkg"}}, contents.ErrInvalidPath},
		{"leading tab", contents.Entry{Path: "\tusr/bin/dpkg", Packages: []string{"admin/dpkg"}}, contents.ErrInvalidPath},
		{"newline", contents.Entry{Path: "usr/bin\n/dpkg", Packages: []string{"admin/dpkg"}}, contents.ErrInvalidPath},
		{"no packages", contents.Entry{Path: "usr/bin/dpkg"}, contents.ErrInvalidPackageList},
		{"comma in package", contents.Entry{Path: "usr/bin/dpkg", Packages: []string{"admin/dpkg,x"}}, contents.ErrInvalidPackageList},
		{"space in package", contents.Entry{Path: "usr/bin/dpkg", Packages: []string{"admin/dpkg x"}}, contents.ErrInvalidPackageList},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.ErrorIs(t, contents.NewWriter(&buf).Write(tc.entry), tc.want)
			require.Empty(t, buf.String())
		})
	}
}

func TestParseQualifiedName(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want contents.QualifiedName
	}{
		{"dpkg", contents.QualifiedName{Name: "dpkg"}},
		{"admin/dpkg", contents.QualifiedName{Section: "admin", Name: "dpkg"}},
		{"contrib/games/crafty", contents.QualifiedName{Area: "contrib", Section: "games", Name: "crafty"}},
		{"universe/x11/xfonts-bolkhov-cp1251-75dpi", contents.QualifiedName{
			Area: "universe", Section: "x11", Name: "xfonts-bolkhov-cp1251-75dpi",
		}},
		{"a/b/c/d", contents.QualifiedName{Area: "a/b", Section: "c", Name: "d"}},
	} {
		t.Run(tc.in, func(t *testing.T) {
			q := contents.ParseQualifiedName(tc.in)
			require.Equal(t, tc.want, q)
			require.Equal(t, tc.in, q.String())
		})
	}
}

func readFile(t *testing.T, path string) []contents.Entry {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	entries, err := contents.NewReader(f).ReadAll()
	require.NoError(t, err)

	return entries
}

func TestReadReturnsEOF(t *testing.T) {
	r := contents.NewReader(strings.NewReader("usr/bin/dpkg admin/dpkg\n"))

	_, err := r.Read()
	require.NoError(t, err)

	_, err = r.Read()
	require.ErrorIs(t, err, io.EOF)
}
