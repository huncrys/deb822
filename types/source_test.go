// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package types_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/boolean"
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/version"
)

// sourceStanza is a Sources index entry for 0ad: the data of the testdata .dsc
// plus the bookkeeping an archive adds to it (Priority, Section, Directory,
// Extra-Source-Only).
const sourceStanza = `Package: 0ad
Format: 3.0 (quilt)
Binary: 0ad
Architecture: amd64 arm64 armhf i386 kfreebsd-amd64 kfreebsd-i386
Version: 0.0.26-3
Priority: optional
Section: games
Maintainer: Debian Games Team <pkg-games-devel@lists.alioth.debian.org>
Uploaders:  Vincent Cheng <vcheng@debian.org>, Ludovic Rousseau <rousseau@debian.org>
Standards-Version: 4.6.2
Build-Depends: autoconf, automake, cmake, debhelper-compat (= 13), libcurl4-gnutls-dev (>= 7.32.0) | libcurl4-dev (>= 7.32.0)
Build-Depends-Indep: fonts-dejavu-core
Testsuite: autopkgtest
Homepage: https://play0ad.com/
Vcs-Browser: https://salsa.debian.org/games-team/0ad
Vcs-Git: https://salsa.debian.org/games-team/0ad.git
Extra-Source-Only: yes
Directory: pool/main/0/0ad
Package-List:
 0ad deb games optional arch=amd64,arm64,armhf,i386,kfreebsd-amd64,kfreebsd-i386
Files:
 7b1a2bd8e2e9a2f9e6a48e6c1a5e0d7b 2565 0ad_0.0.26-3.dsc
 11b79970197c19241708e2a6cadb416d 78065537 0ad_0.0.26.orig.tar.gz
 ef7590961dc6e47d913d9bcec038f52e 5078552 0ad_0.0.26-3.debian.tar.xz
Checksums-Sha256:
 8d0a0f0c1f7a3e0e4d1c9b6a5f2e3d4c5b6a7988776655443322110099887766 2565 0ad_0.0.26-3.dsc
 4a9905004e220d774ff07fd31fe5caab3ada3807eeb7bf664b2904583711421c 78065537 0ad_0.0.26.orig.tar.gz
 2efd0a143ce83496c8984ed3b3e20f2ab84dbc391fcf3d02229d1f1053a1b75c 5078552 0ad_0.0.26-3.debian.tar.xz
`

func encodeSource(t *testing.T, source types.Source) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, deb822.Marshal(&buf, source))

	return buf.Bytes()
}

func TestSource(t *testing.T) {
	var source types.Source
	require.NoError(t, deb822.Unmarshal([]byte(sourceStanza), &source))

	t.Run("scalars", func(t *testing.T) {
		require.Equal(t, "0ad", source.Package)
		require.Equal(t, "3.0 (quilt)", source.Format)
		require.Equal(t, version.MustParse("0.0.26-3"), source.Version)
		require.Equal(t, "optional", source.Priority)
		require.Equal(t, "games", source.Section)
		require.Equal(t, "Debian Games Team <pkg-games-devel@lists.alioth.debian.org>", source.Maintainer)
		require.Equal(t, "4.6.2", source.StandardsVersion)
		require.Equal(t, "autopkgtest", source.Testsuite)
		require.Equal(t, "https://play0ad.com/", source.Homepage)
		require.Equal(t, "https://salsa.debian.org/games-team/0ad", source.VcsBrowser)
		require.Equal(t, "https://salsa.debian.org/games-team/0ad.git", source.VcsGit)
		require.Equal(t, "pool/main/0/0ad", source.Directory)

		require.NotNil(t, source.ExtraSourceOnly)
		require.Equal(t, boolean.Boolean(true), *source.ExtraSourceOnly)

		// Fields the stanza never mentions stay at their zero value.
		require.Empty(t, source.OriginalMaintainer)
		require.Empty(t, source.Description)
		require.Empty(t, source.VcsSvn)
	})

	t.Run("lists", func(t *testing.T) {
		require.Equal(t, list.CommaDelimited[string]{"0ad"}, source.Binary)
		require.Equal(t, list.CommaDelimited[string]{
			"Vincent Cheng <vcheng@debian.org>",
			"Ludovic Rousseau <rousseau@debian.org>",
		}, source.Uploaders)
		require.Equal(t, list.SpaceDelimited[arch.Arch]{
			arch.MustParse("amd64"),
			arch.MustParse("arm64"),
			arch.MustParse("armhf"),
			arch.MustParse("i386"),
			arch.MustParse("kfreebsd-amd64"),
			arch.MustParse("kfreebsd-i386"),
		}, source.Architecture)
		require.Equal(t, list.NewLineDelimited[string]{
			"0ad deb games optional arch=amd64,arm64,armhf,i386,kfreebsd-amd64,kfreebsd-i386",
		}, source.PackageList)
	})

	t.Run("dependencies", func(t *testing.T) {
		require.Len(t, source.BuildDepends.Relations, 5)
		require.Len(t, source.BuildDepends.Relations[4].Possibilities, 2)
		require.Len(t, source.BuildDependsIndep.Relations, 1)
		require.Empty(t, source.BuildDependsArch.Relations)
		require.Empty(t, source.BuildConflicts.Relations)
	})

	t.Run("checksums", func(t *testing.T) {
		require.Len(t, source.Files, 3)
		require.Len(t, source.ChecksumsSha256, 3)
		require.Empty(t, source.ChecksumsSha1)
		require.Empty(t, source.ChecksumsSha512)

		require.Equal(t, filehash.FileHash{
			Hash:     "7b1a2bd8e2e9a2f9e6a48e6c1a5e0d7b",
			Size:     2565,
			Filename: "0ad_0.0.26-3.dsc",
		}, source.Files[0])
		require.Equal(t, filehash.FileHash{
			Hash:     "2efd0a143ce83496c8984ed3b3e20f2ab84dbc391fcf3d02229d1f1053a1b75c",
			Size:     5078552,
			Filename: "0ad_0.0.26-3.debian.tar.xz",
		}, source.ChecksumsSha256[2])
	})

	t.Run("round trip is byte stable", func(t *testing.T) {
		first := encodeSource(t, source)

		var again types.Source
		require.NoError(t, deb822.Unmarshal(first, &again))
		require.Equal(t, source, again, "decoding the encoded stanza changed the value")

		second := encodeSource(t, again)
		require.Equal(t, string(first), string(second), "re-encoding changed the output bytes")
	})

	t.Run("optional fields are left out", func(t *testing.T) {
		encoded := string(encodeSource(t, source))

		require.Contains(t, encoded, "Extra-Source-Only: yes\n")
		require.Contains(t, encoded, "Directory: pool/main/0/0ad\n")
		require.NotContains(t, encoded, "Vcs-Svn")
		require.NotContains(t, encoded, "Original-Maintainer")
		require.NotContains(t, encoded, "Checksums-Sha512")
	})
}
