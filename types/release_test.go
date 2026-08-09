// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package types_test

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	stdtime "time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/boolean"
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/time"
)

func TestRelease(t *testing.T) {
	f, err := os.Open("../testdata/InRelease")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	keyringFile, err := os.Open("../testdata/archive-key-12.asc")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, keyringFile.Close())
	})

	keyring, err := openpgp.ReadArmoredKeyRing(keyringFile)
	require.NoError(t, err)

	decoder, err := deb822.NewDecoder(f, keyring)
	require.NoError(t, err)

	var release types.Release
	require.NoError(t, decoder.Decode(&release))

	require.Equal(t, "Debian", release.Origin)
	require.Equal(t, "Debian", release.Label)
	require.Equal(t, "stable", release.Suite)
	require.Equal(t, "12.5", release.Version)
	require.Equal(t, "bookworm", release.Codename)
	require.Equal(t, "https://metadata.ftp-master.debian.org/changelogs/@CHANGEPATH@_changelog", release.Changelogs)
	require.Equal(t, time.Time(stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC)), release.Date)
	require.Equal(t, boolean.Boolean(true), *release.AcquireByHash)
	require.Equal(t, "Packages", release.NoSupportForArchitectureAll)
	require.Equal(
		t,
		list.SpaceDelimited[arch.Arch]{
			arch.MustParse("all"),
			arch.MustParse("amd64"),
			arch.MustParse("arm64"),
			arch.MustParse("armel"),
			arch.MustParse("armhf"),
			arch.MustParse("i386"),
			arch.MustParse("mips64el"),
			arch.MustParse("mipsel"),
			arch.MustParse("ppc64el"),
			arch.MustParse("s390x"),
		},
		release.Architectures,
	)
	require.Equal(t, list.SpaceDelimited[string]{"main", "contrib", "non-free-firmware", "non-free"}, release.Components)
	require.Equal(t, "Debian 12.5 Released 10 February 2024", release.Description)
	require.Len(t, release.MD5Sum, 772)
	require.Equal(t, filehash.FileHash{
		Hash:     "0ed6d4c8891eb86358b94bb35d9e4da4",
		Size:     1484322,
		Filename: "contrib/Contents-all",
	}, release.MD5Sum[0])
	require.Equal(t, filehash.FileHash{
		Hash:     "d0a0325a97c42fd5f66a8c3e29bcea64",
		Size:     98581,
		Filename: "contrib/Contents-all.gz",
	}, release.MD5Sum[1])
	require.Len(t, release.SHA256, 772)
	require.Equal(t, filehash.FileHash{
		Hash:     "d6c9c82f4e61b4662f9ba16b9ebb379c57b4943f8b7813091d1f637325ddfb79",
		Size:     1484322,
		Filename: "contrib/Contents-all",
	}, release.SHA256[0])
	require.Equal(t, filehash.FileHash{
		Hash:     "c22d03bdd4c7619e1e39e73b4a7b9dfdf1cc1141ed9b10913fbcac58b3a943d0",
		Size:     98581,
		Filename: "contrib/Contents-all.gz",
	}, release.SHA256[1])
}

// encodeRelease renders a release the way the deb822 encoder would.
func encodeRelease(t *testing.T, release types.Release) string {
	t.Helper()

	builder := &strings.Builder{}

	encoder, err := deb822.NewEncoder(builder, nil)
	require.NoError(t, err)

	require.NoError(t, encoder.Encode(release))

	return builder.String()
}

// TestReleaseSHA512RoundTrip covers the SHA512 checksum list, which the sample
// Release file does not carry.
func TestReleaseSHA512RoundTrip(t *testing.T) {
	release := types.Release{
		Origin:        "Debian",
		Suite:         "stable",
		Codename:      "bookworm",
		Date:          time.Time(stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC)),
		Architectures: list.SpaceDelimited[arch.Arch]{arch.MustParse("amd64")},
		Components:    list.SpaceDelimited[string]{"main"},
		SHA512: list.NewLineDelimited[filehash.FileHash]{
			{
				Hash:     "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
				Size:     1484322,
				Filename: "contrib/Contents-all",
			},
			{
				Hash:     "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
				Size:     98581,
				Filename: "contrib/Contents-all.gz",
			},
		},
	}

	encoded := encodeRelease(t, release)
	require.Contains(t, encoded, "SHA512:")
	require.Contains(t, encoded, " cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e 1484322 contrib/Contents-all\n")

	decoder, err := deb822.NewDecoder(strings.NewReader(encoded), nil)
	require.NoError(t, err)

	var decoded types.Release
	require.NoError(t, decoder.Decode(&decoded))

	require.Equal(t, release, decoded)
}

// TestReleaseNoSupportForArchitectureAllSpelling pins the field name spelled
// out by the repository format specification, which lower cases the "for".
func TestReleaseNoSupportForArchitectureAllSpelling(t *testing.T) {
	release := types.Release{
		Origin:                      "Debian",
		Suite:                       "stable",
		Codename:                    "bookworm",
		Date:                        time.Time(stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC)),
		Architectures:               list.SpaceDelimited[arch.Arch]{arch.MustParse("amd64")},
		Components:                  list.SpaceDelimited[string]{"main"},
		NoSupportForArchitectureAll: "Packages",
	}

	encoded := encodeRelease(t, release)

	require.Contains(t, encoded, "No-Support-for-Architecture-all: Packages\n")
	require.NotContains(t, encoded, "No-Support-For-Architecture-all")

	// Decoding is case insensitive, so either spelling still parses.
	for _, key := range []string{"No-Support-for-Architecture-all", "No-Support-For-Architecture-all"} {
		decoder, err := deb822.NewDecoder(strings.NewReader(key+": Packages\n"), nil)
		require.NoError(t, err)

		var decoded types.Release
		require.NoError(t, decoder.Decode(&decoded))
		require.Equal(t, "Packages", decoded.NoSupportForArchitectureAll, "failed to decode %s", key)
	}
}

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func TestSums(t *testing.T) {
	release := types.Release{
		MD5Sum: list.NewLineDelimited[filehash.FileHash]{
			{
				Hash:     "0ed6d4c8891eb86358b94bb35d9e4da4",
				Size:     1484322,
				Filename: "contrib/Contents-all",
			},
			{
				Hash:     "d0a0325a97c42fd5f66a8c3e29bcea64",
				Size:     98581,
				Filename: "contrib/Contents-all.gz",
			},
		},
		SHA1: list.NewLineDelimited[filehash.FileHash]{
			{
				Hash:     "3b5d5c3712955042212316173ccf37be800a6f3f",
				Size:     1484322,
				Filename: "contrib/Contents-all",
			},
			{
				Hash:     "5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8",
				Size:     98581,
				Filename: "contrib/Contents-all.gz",
			},
		},
		SHA256: list.NewLineDelimited[filehash.FileHash]{
			{
				Hash:     "d6c9c82f4e61b4662f9ba16b9ebb379c57b4943f8b7813091d1f637325ddfb79",
				Size:     1484322,
				Filename: "contrib/Contents-all",
			},
			{
				Hash:     "c22d03bdd4c7619e1e39e73b4a7b9dfdf1cc1141ed9b10913fbcac58b3a943d0",
				Size:     98581,
				Filename: "contrib/Contents-all.gz",
			},
		},
		SHA512: list.NewLineDelimited[filehash.FileHash]{
			{
				Hash:     "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
				Size:     1484322,
				Filename: "contrib/Contents-all",
			},
			{
				Hash:     "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
				Size:     98581,
				Filename: "contrib/Contents-all.gz",
			},
		},
	}

	expected := map[string][]byte{
		"contrib/Contents-all":    mustDecodeHex(t, "0ed6d4c8891eb86358b94bb35d9e4da4"),
		"contrib/Contents-all.gz": mustDecodeHex(t, "d0a0325a97c42fd5f66a8c3e29bcea64"),
	}

	sums, err := release.MD5Sums()
	require.NoError(t, err)

	require.Len(t, sums, 2)
	require.Equal(t, expected, sums)

	expected = map[string][]byte{
		"contrib/Contents-all":    mustDecodeHex(t, "3b5d5c3712955042212316173ccf37be800a6f3f"),
		"contrib/Contents-all.gz": mustDecodeHex(t, "5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8"),
	}

	sums, err = release.SHA1Sums()
	require.NoError(t, err)

	require.Len(t, sums, 2)
	require.Equal(t, expected, sums)

	expected = map[string][]byte{
		"contrib/Contents-all":    mustDecodeHex(t, "d6c9c82f4e61b4662f9ba16b9ebb379c57b4943f8b7813091d1f637325ddfb79"),
		"contrib/Contents-all.gz": mustDecodeHex(t, "c22d03bdd4c7619e1e39e73b4a7b9dfdf1cc1141ed9b10913fbcac58b3a943d0"),
	}

	sums, err = release.SHA256Sums()
	require.NoError(t, err)

	require.Len(t, sums, 2)
	require.Equal(t, expected, sums)

	expected = map[string][]byte{
		"contrib/Contents-all":    mustDecodeHex(t, "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"),
		"contrib/Contents-all.gz": mustDecodeHex(t, "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"),
	}

	sums, err = release.SHA512Sums()
	require.NoError(t, err)

	require.Len(t, sums, 2)
	require.Equal(t, expected, sums)

	t.Run("invalid hash", func(t *testing.T) {
		release := types.Release{
			MD5Sum: list.NewLineDelimited[filehash.FileHash]{
				{
					Hash:     "invalidhash",
					Size:     123,
					Filename: "file.txt",
				},
			},
		}

		_, err := release.MD5Sums()
		require.Error(t, err)
	})
}
