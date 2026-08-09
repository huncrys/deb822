// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package deb822_test

import (
	"compress/gzip"
	"os"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
)

func TestStrictRejects(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{
			name:  "space before colon",
			input: "Key : value\n",
			want:  deb822.ErrInvalidFieldName,
		},
		{
			name:  "leading hyphen",
			input: "-Key: value\n",
			want:  deb822.ErrInvalidFieldName,
		},
		{
			name:  "non ascii character",
			input: "K\xc3\xa9y: value\n",
			want:  deb822.ErrInvalidFieldName,
		},
		{
			name:  "empty name",
			input: ": value\n",
			want:  deb822.ErrInvalidFieldName,
		},
		{
			name:  "duplicate field",
			input: "Foo: one\nFoo: two\n",
			want:  deb822.ErrDuplicateField,
		},
		{
			name:  "duplicate field different case",
			input: "Foo: one\nFOO: two\n",
			want:  deb822.ErrDuplicateField,
		},
		{
			name:  "comment line",
			input: "Foo: one\n# comment\nBar: two\n",
			want:  deb822.ErrCommentNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := deb822.NewStanzaReader(strings.NewReader(tt.input), nil, deb822.WithStrict())
			require.NoError(t, err)

			_, err = reader.All()
			require.Error(t, err)
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestStrictAccepts(t *testing.T) {
	reader, err := deb822.NewStanzaReader(strings.NewReader(`Package: hello
Description: a greeting
 more text
 .
 and more

Package: goodbye
`), nil, deb822.WithStrict())
	require.NoError(t, err)

	blocks, err := reader.All()
	require.NoError(t, err)

	require.Len(t, blocks, 2)
	require.Equal(t, "hello", blocks[0].Values["Package"])
	require.Equal(t, "a greeting\nmore text\n\nand more\n", blocks[0].Values["Description"])
	require.Equal(t, "goodbye", blocks[1].Values["Package"])
}

func TestCommentPolicy(t *testing.T) {
	const withComment = "Foo: one\n# comment\nBar: two\n"

	t.Run("lenient default allows comments", func(t *testing.T) {
		reader, err := deb822.NewStanzaReader(strings.NewReader(withComment), nil)
		require.NoError(t, err)

		blocks, err := reader.All()
		require.NoError(t, err)
		require.Len(t, blocks, 1)
	})

	t.Run("lenient with comments disabled", func(t *testing.T) {
		reader, err := deb822.NewStanzaReader(strings.NewReader(withComment), nil, deb822.WithComments(false))
		require.NoError(t, err)

		_, err = reader.All()
		require.ErrorIs(t, err, deb822.ErrCommentNotAllowed)
	})

	t.Run("strict with comments enabled", func(t *testing.T) {
		reader, err := deb822.NewStanzaReader(
			strings.NewReader(withComment), nil, deb822.WithStrict(), deb822.WithComments(true),
		)
		require.NoError(t, err)

		blocks, err := reader.All()
		require.NoError(t, err)
		require.Len(t, blocks, 1)
		require.Equal(t, "one", blocks[0].Values["Foo"])
		require.Equal(t, "two", blocks[0].Values["Bar"])
	})

	t.Run("strict with comments enabled still checks names", func(t *testing.T) {
		reader, err := deb822.NewStanzaReader(
			strings.NewReader("# comment\nFoo : one\n"), nil, deb822.WithStrict(), deb822.WithComments(true),
		)
		require.NoError(t, err)

		_, err = reader.All()
		require.ErrorIs(t, err, deb822.ErrInvalidFieldName)
	})

	t.Run("strict with comments enabled still checks duplicates", func(t *testing.T) {
		reader, err := deb822.NewStanzaReader(
			strings.NewReader("# comment\nFoo: one\nfoo: two\n"), nil, deb822.WithStrict(), deb822.WithComments(true),
		)
		require.NoError(t, err)

		_, err = reader.All()
		require.ErrorIs(t, err, deb822.ErrDuplicateField)
	})

	t.Run("option order does not matter", func(t *testing.T) {
		reader, err := deb822.NewStanzaReader(
			strings.NewReader(withComment), nil, deb822.WithComments(true), deb822.WithStrict(),
		)
		require.NoError(t, err)

		blocks, err := reader.All()
		require.NoError(t, err)
		require.Len(t, blocks, 1)
	})
}

func TestUnexpectedContinuation(t *testing.T) {
	tests := []struct {
		name string
		opts []deb822.ReaderOption
	}{
		{name: "lenient"},
		{name: "strict", opts: []deb822.ReaderOption{deb822.WithStrict()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				reader, err := deb822.NewStanzaReader(strings.NewReader(" foo\nKey: value\n"), nil, tt.opts...)
				require.NoError(t, err)

				_, err = reader.All()
				require.ErrorIs(t, err, deb822.ErrUnexpectedContinuation)
			})
		})
	}

	t.Run("continuation after blank separator", func(t *testing.T) {
		require.NotPanics(t, func() {
			reader, err := deb822.NewStanzaReader(strings.NewReader("Key: value\n\n continuation\n"), nil)
			require.NoError(t, err)

			_, err = reader.All()
			require.ErrorIs(t, err, deb822.ErrUnexpectedContinuation)
		})
	})
}

func TestUnmarshalWithOptions(t *testing.T) {
	type testStruct struct {
		Foo string
	}

	var lenient testStruct
	require.NoError(t, deb822.Unmarshal([]byte("Foo : one\n"), &lenient))
	require.Equal(t, "one", lenient.Foo)

	var strict testStruct
	err := deb822.Unmarshal([]byte("Foo : one\n"), &strict, deb822.WithStrict())
	require.ErrorIs(t, err, deb822.ErrInvalidFieldName)
}

// TestLenientDefaultsUnchanged guards the default (no options) behaviour that
// downstream consumers depend on.
func TestLenientDefaultsUnchanged(t *testing.T) {
	t.Run("Packages.gz", func(t *testing.T) {
		f, err := os.Open("testdata/Packages.gz")
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, f.Close())
		})

		dr, err := gzip.NewReader(f)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, dr.Close())
		})

		reader, err := deb822.NewStanzaReader(dr, nil)
		require.NoError(t, err)

		blocks, err := reader.All()
		require.NoError(t, err)

		require.Len(t, blocks, 63408)
		require.Equal(t, "0ad", blocks[0].Values["Package"])
	})

	t.Run("InRelease", func(t *testing.T) {
		f, err := os.Open("testdata/InRelease")
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, f.Close())
		})

		keyringFile, err := os.Open("testdata/archive-key-12.asc")
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, keyringFile.Close())
		})

		keyring, err := openpgp.ReadArmoredKeyRing(keyringFile)
		require.NoError(t, err)

		reader, err := deb822.NewStanzaReader(f, keyring)
		require.NoError(t, err)
		require.NotNil(t, reader.Signer())

		blocks, err := reader.All()
		require.NoError(t, err)

		require.Len(t, blocks, 1)
		require.Equal(t, "Debian", blocks[0].Values["Origin"])
	})
}

// TestNilOptionIgnored makes sure a nil option in the variadic list is a no-op
// rather than a panic.
func TestNilOptionIgnored(t *testing.T) {
	require.NotPanics(t, func() {
		_, err := deb822.NewStanzaReader(strings.NewReader("Foo: one\n"), nil, nil)
		require.NoError(t, err)
	})
}
