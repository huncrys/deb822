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
	"os"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types"
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
}
