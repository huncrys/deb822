// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package filehash_test

import (
	"os"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
)

func TestFileHash(t *testing.T) {
	f, err := os.Open("../../testdata/InRelease")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	keyringFile, err := os.Open("../../testdata/archive-key-12.asc")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, keyringFile.Close())
	})

	keyring, err := openpgp.ReadArmoredKeyRing(keyringFile)
	require.NoError(t, err)

	decoder, err := deb822.NewDecoder(f, keyring)
	require.NoError(t, err)

	type TestStruct struct {
		MD5Sum list.NewLineDelimited[filehash.FileHash]
		SHA256 list.NewLineDelimited[filehash.FileHash]
	}

	var foo TestStruct
	require.NoError(t, decoder.Decode(&foo))

	require.Len(t, foo.MD5Sum, 772)
	require.Len(t, foo.SHA256, 772)

	require.Equal(t, "0ed6d4c8891eb86358b94bb35d9e4da4", foo.MD5Sum[0].Hash)
	require.Equal(t, int64(1484322), foo.MD5Sum[0].Size)
	require.Equal(t, "contrib/Contents-all", foo.MD5Sum[0].Filename)
}

func TestFileHash_MarshalText(t *testing.T) {
	hashes := list.NewLineDelimited[filehash.FileHash]([]filehash.FileHash{{
		Hash:     "0ed6d4c8891eb86358b94bb35d9e4da4",
		Size:     1484322,
		Filename: "contrib/Contents-all",
	}, {
		Hash:     "d0a0325a97c42fd5f66a8c3e29bcea64",
		Size:     98581,
		Filename: "contrib/Contents-all.gz",
	}})

	text, err := hashes.MarshalText()
	require.NoError(t, err)

	expected := `
0ed6d4c8891eb86358b94bb35d9e4da4 1484322 contrib/Contents-all
d0a0325a97c42fd5f66a8c3e29bcea64 98581 contrib/Contents-all.gz`

	require.Equal(t, expected, string(text))
}

func TestFileHash_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected filehash.FileHash
	}{
		{
			name:  "simple",
			input: "0ed6d4c8891eb86358b94bb35d9e4da4 1484322 contrib/Contents-all",
			expected: filehash.FileHash{
				Hash:     "0ed6d4c8891eb86358b94bb35d9e4da4",
				Size:     1484322,
				Filename: "contrib/Contents-all",
			},
		},
		{
			name:  "filename with a single interior space",
			input: "abc123 42 pool/main/some dir/foo.deb",
			expected: filehash.FileHash{
				Hash:     "abc123",
				Size:     42,
				Filename: "pool/main/some dir/foo.deb",
			},
		},
		{
			name:  "filename with multiple interior spaces",
			input: "abc123 42 pool/main/a b/c  d/foo bar.deb",
			expected: filehash.FileHash{
				Hash:     "abc123",
				Size:     42,
				Filename: "pool/main/a b/c  d/foo bar.deb",
			},
		},
		{
			name:  "repeated separator spaces",
			input: "abc123   42    pool/main/foo.deb",
			expected: filehash.FileHash{
				Hash:     "abc123",
				Size:     42,
				Filename: "pool/main/foo.deb",
			},
		},
		{
			name:  "leading separator spaces",
			input: "   abc123 42 pool/main/foo.deb",
			expected: filehash.FileHash{
				Hash:     "abc123",
				Size:     42,
				Filename: "pool/main/foo.deb",
			},
		},
		{
			name:  "repeated separator spaces before a filename with interior spaces",
			input: "abc123  42  a b  c.deb",
			expected: filehash.FileHash{
				Hash:     "abc123",
				Size:     42,
				Filename: "a b  c.deb",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var h filehash.FileHash
			require.NoError(t, h.UnmarshalText([]byte(tc.input)))
			require.Equal(t, tc.expected, h)
		})
	}
}

func TestFileHash_UnmarshalText_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "hash only", input: "abc123"},
		{name: "missing filename", input: "abc123 42"},
		{name: "empty filename", input: "abc123 42 "},
		{name: "non numeric size", input: "abc123 garbage pool/main/foo.deb"},
		{name: "float size", input: "abc123 42.5 pool/main/foo.deb"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var h filehash.FileHash
			err := h.UnmarshalText([]byte(tc.input))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.input)
		})
	}
}

func TestFileHash_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		fileHash filehash.FileHash
	}{
		{
			name: "simple",
			fileHash: filehash.FileHash{
				Hash:     "0ed6d4c8891eb86358b94bb35d9e4da4",
				Size:     1484322,
				Filename: "contrib/Contents-all",
			},
		},
		{
			name: "filename with interior spaces",
			fileHash: filehash.FileHash{
				Hash:     "d0a0325a97c42fd5f66a8c3e29bcea64",
				Size:     98581,
				Filename: "pool/main/a b/c  d/foo bar.deb",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			text, err := tc.fileHash.MarshalText()
			require.NoError(t, err)
			require.Equal(t, tc.fileHash.String(), string(text))

			var got filehash.FileHash
			require.NoError(t, got.UnmarshalText(text))
			require.Equal(t, tc.fileHash, got)
		})
	}
}

func TestChangesFileHash_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected filehash.ChangesFileHash
	}{
		{
			name:  "realistic changes entry",
			input: "abc123 1234 utils optional foo_1.0-1_amd64.deb",
			expected: filehash.ChangesFileHash{
				Hash:     "abc123",
				Size:     1234,
				Section:  "utils",
				Priority: "optional",
				Filename: "foo_1.0-1_amd64.deb",
			},
		},
		{
			name:  "filename with a single interior space",
			input: "abc123 1234 utils optional some dir/foo.deb",
			expected: filehash.ChangesFileHash{
				Hash:     "abc123",
				Size:     1234,
				Section:  "utils",
				Priority: "optional",
				Filename: "some dir/foo.deb",
			},
		},
		{
			name:  "filename with multiple interior spaces",
			input: "abc123 1234 utils optional a b/c  d/foo bar.deb",
			expected: filehash.ChangesFileHash{
				Hash:     "abc123",
				Size:     1234,
				Section:  "utils",
				Priority: "optional",
				Filename: "a b/c  d/foo bar.deb",
			},
		},
		{
			name:  "repeated separator spaces",
			input: "  abc123   1234  utils    optional   a b/foo.deb",
			expected: filehash.ChangesFileHash{
				Hash:     "abc123",
				Size:     1234,
				Section:  "utils",
				Priority: "optional",
				Filename: "a b/foo.deb",
			},
		},
		{
			name:  "section with a slash",
			input: "d41d8cd98f00b204e9800998ecf8427e 0 non-free/libs extra bar_2.0_all.deb",
			expected: filehash.ChangesFileHash{
				Hash:     "d41d8cd98f00b204e9800998ecf8427e",
				Size:     0,
				Section:  "non-free/libs",
				Priority: "extra",
				Filename: "bar_2.0_all.deb",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var h filehash.ChangesFileHash
			require.NoError(t, h.UnmarshalText([]byte(tc.input)))
			require.Equal(t, tc.expected, h)
		})
	}
}

func TestChangesFileHash_UnmarshalText_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "hash only", input: "abc123"},
		{name: "missing section", input: "abc123 1234"},
		{name: "missing priority", input: "abc123 1234 utils"},
		{name: "missing filename", input: "abc123 1234 utils optional"},
		{name: "empty filename", input: "abc123 1234 utils optional "},
		{name: "non numeric size", input: "abc123 garbage utils optional foo.deb"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var h filehash.ChangesFileHash
			err := h.UnmarshalText([]byte(tc.input))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.input)
		})
	}
}

func TestChangesFileHash_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		fileHash filehash.ChangesFileHash
		expected string
	}{
		{
			name: "realistic changes entry",
			fileHash: filehash.ChangesFileHash{
				Hash:     "abc123",
				Size:     1234,
				Section:  "utils",
				Priority: "optional",
				Filename: "foo_1.0-1_amd64.deb",
			},
			expected: "abc123 1234 utils optional foo_1.0-1_amd64.deb",
		},
		{
			name: "filename with interior spaces",
			fileHash: filehash.ChangesFileHash{
				Hash:     "abc123",
				Size:     1234,
				Section:  "utils",
				Priority: "optional",
				Filename: "a b/c  d/foo bar.deb",
			},
			expected: "abc123 1234 utils optional a b/c  d/foo bar.deb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.fileHash.String())

			text, err := tc.fileHash.MarshalText()
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(text))

			var got filehash.ChangesFileHash
			require.NoError(t, got.UnmarshalText(text))
			require.Equal(t, tc.fileHash, got)
		})
	}
}

func TestChangesFileHash_MarshalText_List(t *testing.T) {
	hashes := list.NewLineDelimited[filehash.ChangesFileHash]([]filehash.ChangesFileHash{{
		Hash:     "abc123",
		Size:     1234,
		Section:  "utils",
		Priority: "optional",
		Filename: "foo_1.0-1_amd64.deb",
	}, {
		Hash:     "def456",
		Size:     5678,
		Section:  "utils",
		Priority: "optional",
		Filename: "foo_1.0-1.dsc",
	}})

	text, err := hashes.MarshalText()
	require.NoError(t, err)

	expected := `
abc123 1234 utils optional foo_1.0-1_amd64.deb
def456 5678 utils optional foo_1.0-1.dsc`

	require.Equal(t, expected, string(text))
}
