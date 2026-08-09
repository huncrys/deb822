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

	stdtime "time"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/version"
)

// changesStanza is a source plus binary upload of 0ad. It carries a numeric
// zone Date, the "source" pseudo architecture, a Files field in the five field
// .changes form, and a Changes body whose blank line is written as the lone
// dot placeholder.
const changesStanza = `Format: 1.8
Date: Thu, 07 Mar 2024 12:34:56 +0100
Source: 0ad
Binary: 0ad 0ad-dbgsym
Architecture: source amd64
Version: 0.0.26-3
Distribution: unstable
Urgency: medium
Maintainer: Debian Games Team <pkg-games-devel@lists.alioth.debian.org>
Changed-By: Vincent Cheng <vcheng@debian.org>
Description:
 0ad        - Real-time strategy game of ancient warfare
Closes: 1054321 1060001
Changes:
 0ad (0.0.26-3) unstable; urgency=medium
 .
   * Rebuild against the new libfmt. (Closes: #1054321)
   * Drop the obsolete build profile. (Closes: #1060001)
Checksums-Sha1:
 8e054aa27d9c0e7d1b1c52fc4fa9ee9e230483b7 2565 0ad_0.0.26-3.dsc
 86b98f015e11cc545c8658e66179fc80c6ba12d4 7891488 0ad_0.0.26-3_amd64.deb
Checksums-Sha256:
 4a9905004e220d774ff07fd31fe5caab3ada3807eeb7bf664b2904583711421c 2565 0ad_0.0.26-3.dsc
 3a2118df47bf3f04285649f0455c2fc6fe2dc7f0b237073038aa00af41f0d5f2 7891488 0ad_0.0.26-3_amd64.deb
Files:
 7b1a2bd8e2e9a2f9e6a48e6c1a5e0d7b 2565 games optional 0ad_0.0.26-3.dsc
 4d471183a39a3a11d00cd35bf9f6803d 7891488 games optional 0ad_0.0.26-3_amd64.deb
`

func encodeChanges(t *testing.T, changes types.Changes) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, deb822.Marshal(&buf, changes))

	return buf.Bytes()
}

func TestChanges(t *testing.T) {
	var changes types.Changes
	require.NoError(t, deb822.Unmarshal([]byte(changesStanza), &changes))

	t.Run("scalars", func(t *testing.T) {
		require.Equal(t, "1.8", changes.Format)
		require.Equal(t, "0ad", changes.Source)
		require.Equal(t, version.MustParse("0.0.26-3"), changes.Version)
		require.Equal(t, "medium", changes.Urgency)
		require.Equal(t, "Debian Games Team <pkg-games-devel@lists.alioth.debian.org>", changes.Maintainer)
		require.Equal(t, "Vincent Cheng <vcheng@debian.org>", changes.ChangedBy)
	})

	t.Run("date keeps its numeric zone", func(t *testing.T) {
		parsed := stdtime.Time(changes.Date)

		require.True(t, parsed.Equal(stdtime.Date(2024, stdtime.March, 7, 11, 34, 56, 0, stdtime.UTC)))

		zone, offset := parsed.Zone()
		require.Equal(t, 3600, offset)
		require.Empty(t, zone)

		require.Equal(t, "Thu, 07 Mar 2024 12:34:56 +0100", parsed.Format(stdtime.RFC1123))
	})

	t.Run("binary is space delimited", func(t *testing.T) {
		require.Equal(t, list.SpaceDelimited[string]{"0ad", "0ad-dbgsym"}, changes.Binary)
		require.Equal(t, list.SpaceDelimited[string]{"unstable"}, changes.Distribution)
		require.Equal(t, list.SpaceDelimited[string]{"1054321", "1060001"}, changes.Closes)
	})

	t.Run("source pseudo architecture", func(t *testing.T) {
		require.Len(t, changes.Architecture, 2)

		// "source" is not a real architecture, but it is an unknown single
		// token, so the tuple parser pads it with the usual defaults.
		require.Equal(t, arch.Arch{
			ABI:  "base",
			Libc: "gnu",
			OS:   "linux",
			CPU:  "source",
		}, changes.Architecture[0])
		require.Equal(t, "source", changes.Architecture[0].String())

		require.Equal(t, arch.MustParse("amd64"), changes.Architecture[1])

		text, err := changes.Architecture.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "source amd64", string(text))
	})

	t.Run("multiline bodies", func(t *testing.T) {
		require.Equal(t, "0ad        - Real-time strategy game of ancient warfare\n", changes.Description)
		require.Equal(t, "0ad (0.0.26-3) unstable; urgency=medium\n"+
			"\n"+
			"  * Rebuild against the new libfmt. (Closes: #1054321)\n"+
			"  * Drop the obsolete build profile. (Closes: #1060001)\n",
			changes.Changes)
	})

	t.Run("checksums", func(t *testing.T) {
		require.Len(t, changes.ChecksumsSha1, 2)
		require.Len(t, changes.ChecksumsSha256, 2)
		require.Len(t, changes.Files, 2)

		require.Equal(t, filehash.ChangesFileHash{
			Hash:     "7b1a2bd8e2e9a2f9e6a48e6c1a5e0d7b",
			Size:     2565,
			Section:  "games",
			Priority: "optional",
			Filename: "0ad_0.0.26-3.dsc",
		}, changes.Files[0])
		require.Equal(t, filehash.ChangesFileHash{
			Hash:     "4d471183a39a3a11d00cd35bf9f6803d",
			Size:     7891488,
			Section:  "games",
			Priority: "optional",
			Filename: "0ad_0.0.26-3_amd64.deb",
		}, changes.Files[1])
		require.Equal(t, filehash.FileHash{
			Hash:     "8e054aa27d9c0e7d1b1c52fc4fa9ee9e230483b7",
			Size:     2565,
			Filename: "0ad_0.0.26-3.dsc",
		}, changes.ChecksumsSha1[0])
	})

	t.Run("round trip is byte stable", func(t *testing.T) {
		first := encodeChanges(t, changes)

		// The dot placeholder survives the trip: the blank line of the Changes
		// body is written back out as " .".
		require.Contains(t, string(first), "\n .\n")
		require.Contains(t, string(first), "Architecture: source amd64\n")
		require.Contains(t, string(first), "Date: Thu, 07 Mar 2024 12:34:56 +0100\n")

		var again types.Changes
		require.NoError(t, deb822.Unmarshal(first, &again))

		second := encodeChanges(t, again)
		require.Equal(t, string(first), string(second), "re-encoding changed the output bytes")
	})
}
