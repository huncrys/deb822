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
	"os"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/version"
)

// TestDsc decodes the clearsigned .dsc of 0ad 0.0.26-3, which exercises the
// whole document: a signed stanza, comma and space delimited lists, a long
// Build-Depends with an alternative, and three multi line file lists.
func TestDsc(t *testing.T) {
	f, err := os.Open("../testdata/0ad_0.0.26-3.dsc")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	keyringFile, err := os.Open("../testdata/d53a815a3cb7659af882e3958eedcc1baa1f32ff.asc")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, keyringFile.Close())
	})

	keyring, err := openpgp.ReadArmoredKeyRing(keyringFile)
	require.NoError(t, err)

	decoder, err := deb822.NewDecoder(f, keyring)
	require.NoError(t, err)

	var dsc types.Dsc
	require.NoError(t, decoder.Decode(&dsc))

	require.NotNil(t, decoder.Signer(), "the clearsigned .dsc should resolve to a signer")

	t.Run("scalars", func(t *testing.T) {
		require.Equal(t, "3.0 (quilt)", dsc.Format)
		require.Equal(t, "0ad", dsc.Source)
		require.Equal(t, version.MustParse("0.0.26-3"), dsc.Version)
		require.Equal(t, "Debian Games Team <pkg-games-devel@lists.alioth.debian.org>", dsc.Maintainer)
		require.Equal(t, "https://play0ad.com/", dsc.Homepage)
		require.Equal(t, "4.6.2", dsc.StandardsVersion)
		require.Equal(t, "https://salsa.debian.org/games-team/0ad", dsc.VcsBrowser)
		require.Equal(t, "https://salsa.debian.org/games-team/0ad.git", dsc.VcsGit)
	})

	t.Run("lists", func(t *testing.T) {
		require.Equal(t, list.CommaDelimited[string]{"0ad"}, dsc.Binary)
		require.Equal(t, list.CommaDelimited[string]{
			"Vincent Cheng <vcheng@debian.org>",
			"Ludovic Rousseau <rousseau@debian.org>",
		}, dsc.Uploaders)
		require.Equal(t, list.SpaceDelimited[arch.Arch]{
			arch.MustParse("amd64"),
			arch.MustParse("arm64"),
			arch.MustParse("armhf"),
			arch.MustParse("i386"),
			arch.MustParse("kfreebsd-amd64"),
			arch.MustParse("kfreebsd-i386"),
		}, dsc.Architecture)
	})

	t.Run("build depends", func(t *testing.T) {
		// The field spells out 32 comma separated relations, one of which
		// (libcurl4-gnutls-dev | libcurl4-dev) offers two possibilities.
		require.Len(t, dsc.BuildDepends.Relations, 32)
		require.Empty(t, dsc.BuildDependsIndep.Relations)
		require.Empty(t, dsc.BuildDependsArch.Relations)
		require.Empty(t, dsc.BuildConflicts.Relations)

		require.Equal(t, "autoconf", dsc.BuildDepends.Relations[0].String())
		require.Equal(t, "debhelper-compat (= 13)", dsc.BuildDepends.Relations[4].String())

		var alternatives int
		for _, relation := range dsc.BuildDepends.Relations {
			if len(relation.Possibilities) > 1 {
				alternatives++
			}
		}
		require.Equal(t, 1, alternatives)
	})

	t.Run("package list", func(t *testing.T) {
		require.Equal(t, list.NewLineDelimited[string]{
			"0ad deb games optional arch=amd64,arm64,armhf,i386,kfreebsd-amd64,kfreebsd-i386",
		}, dsc.PackageList)
	})

	t.Run("checksums", func(t *testing.T) {
		require.Len(t, dsc.Files, 2)
		require.Len(t, dsc.ChecksumsSha1, 2)
		require.Len(t, dsc.ChecksumsSha256, 2)
		require.Empty(t, dsc.ChecksumsSha512)

		require.Equal(t, filehash.FileHash{
			Hash:     "11b79970197c19241708e2a6cadb416d",
			Size:     78065537,
			Filename: "0ad_0.0.26.orig.tar.gz",
		}, dsc.Files[0])
		require.Equal(t, filehash.FileHash{
			Hash:     "4a9905004e220d774ff07fd31fe5caab3ada3807eeb7bf664b2904583711421c",
			Size:     78065537,
			Filename: "0ad_0.0.26.orig.tar.gz",
		}, dsc.ChecksumsSha256[0])
		require.Equal(t, filehash.FileHash{
			Hash:     "2efd0a143ce83496c8984ed3b3e20f2ab84dbc391fcf3d02229d1f1053a1b75c",
			Size:     5078552,
			Filename: "0ad_0.0.26-3.debian.tar.xz",
		}, dsc.ChecksumsSha256[1])
	})
}
