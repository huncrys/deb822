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
