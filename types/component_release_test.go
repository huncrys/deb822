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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/boolean"
)

// debianComponentRelease is dists/trixie/main/binary-amd64/Release as
// deb.debian.org serves it. It is the ground truth for the field order.
const debianComponentRelease = `Archive: stable
Origin: Debian
Label: Debian
Version: 13.6
Acquire-By-Hash: yes
Component: main
Architecture: amd64
`

// TestComponentRelease pins the serialized form of the stub against the file
// Debian publishes, field for field and in order.
func TestComponentRelease(t *testing.T) {
	acquireByHash := boolean.Boolean(true)

	stub := types.ComponentRelease{
		Archive:       "stable",
		Origin:        "Debian",
		Label:         "Debian",
		Version:       "13.6",
		AcquireByHash: &acquireByHash,
		Component:     "main",
		Architecture:  arch.MustParse("amd64"),
	}

	builder := &strings.Builder{}

	encoder, err := deb822.NewEncoder(builder, nil)
	require.NoError(t, err)

	require.NoError(t, encoder.Encode(stub))
	require.Equal(t, debianComponentRelease, builder.String())

	decoder, err := deb822.NewDecoder(strings.NewReader(debianComponentRelease), nil)
	require.NoError(t, err)

	var decoded types.ComponentRelease
	require.NoError(t, decoder.Decode(&decoded))

	require.Equal(t, stub, decoded)
}

// TestComponentReleaseOptionalFieldsAreOmitted pins the shape of the minimal
// stub: a repository that names neither an origin nor a version must not grow
// empty fields for them.
func TestComponentReleaseOptionalFieldsAreOmitted(t *testing.T) {
	stub := types.ComponentRelease{
		Archive:      "nodistro",
		Component:    "main",
		Architecture: arch.MustParse("arm64"),
	}

	builder := &strings.Builder{}

	encoder, err := deb822.NewEncoder(builder, nil)
	require.NoError(t, err)

	require.NoError(t, encoder.Encode(stub))

	require.Equal(t, `Archive: nodistro
Component: main
Architecture: arm64
`, builder.String())
}

// TestComponentReleaseEncodeIsStable pins that a decode/encode cycle of the
// stub reaches a fixed point, the way TestEncodeIsStable does for packages.
func TestComponentReleaseEncodeIsStable(t *testing.T) {
	decoder, err := deb822.NewDecoder(strings.NewReader(debianComponentRelease), nil)
	require.NoError(t, err)

	var decoded types.ComponentRelease
	require.NoError(t, decoder.Decode(&decoded))

	builder := &strings.Builder{}

	encoder, err := deb822.NewEncoder(builder, nil)
	require.NoError(t, err)

	require.NoError(t, encoder.Encode(decoded))
	require.Equal(t, debianComponentRelease, builder.String())
}
