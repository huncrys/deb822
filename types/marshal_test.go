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
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types"
)

// samplePackages decodes the head of the testdata Packages file.
func samplePackages(t *testing.T) []types.Package {
	t.Helper()

	f, err := os.Open("../testdata/Packages.gz")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	dr, err := gzip.NewReader(f)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dr.Close())
	})

	decoder, err := deb822.NewDecoder(io.LimitReader(dr, 1000000), nil)
	require.NoError(t, err)

	var packages []types.Package
	require.NoError(t, decoder.Decode(&packages))
	require.NotEmpty(t, packages)

	// The limited read cuts the file mid-stanza, so the last entry is a
	// partial package. Drop it, the tests below want whole ones.
	return packages[:len(packages)-1]
}

func encodePackages(t *testing.T, packages []types.Package) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, deb822.Marshal(&buf, packages))

	return buf.Bytes()
}

// TestEncodeIsStable pins that a decode/encode cycle reaches a fixed point:
// whatever the first encode produces, decoding and encoding it again produces
// the very same bytes. It is the self contained stand-in for a byte for byte
// comparison against a captured baseline.
func TestEncodeIsStable(t *testing.T) {
	first := encodePackages(t, samplePackages(t))

	decoder, err := deb822.NewDecoder(bytes.NewReader(first), nil)
	require.NoError(t, err)

	var packages []types.Package
	require.NoError(t, decoder.Decode(&packages))

	second := encodePackages(t, packages)

	require.Equal(t, len(first), len(second), "re-encoding changed the output length")
	require.True(t, bytes.Equal(first, second), "re-encoding changed the output bytes")
}

// TestDecodeIsCaseInsensitive covers Debian Policy 5.1: field names are
// matched without regard to case.
func TestDecodeIsCaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "as declared", input: "Package: foo\nVersion: 1.0-1\n"},
		{name: "lower case", input: "package: foo\nversion: 1.0-1\n"},
		{name: "upper case", input: "PACKAGE: foo\nVERSION: 1.0-1\n"},
		{name: "mixed case", input: "packAGE: foo\nvErSiOn: 1.0-1\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var pkg types.Package
			require.NoError(t, deb822.Unmarshal([]byte(test.input), &pkg))

			require.Equal(t, "foo", pkg.Name)
			require.Equal(t, "1.0-1", pkg.Version.String())
		})
	}
}

// TestJSONHasNoEmptyValues guards the json tags of Package: a plain
// json.Marshal of a decoded package must not carry the fields the package
// never set. The stanza path used to do that filtering, so a consumer dumping
// a Package as JSON got a wall of empty strings.
func TestJSONHasNoEmptyValues(t *testing.T) {
	packages := samplePackages(t)

	t.Run("sparse package", func(t *testing.T) {
		// The first entry sets a good handful of fields but leaves most of the
		// optional ones alone.
		encoded, err := json.Marshal(packages[0])
		require.NoError(t, err)

		require.NotContains(t, string(encoded), `":""`, "empty string value in %s", encoded)
		require.Contains(t, string(encoded), `"Package":"0ad"`)
	})

	t.Run("every package", func(t *testing.T) {
		for _, pkg := range packages {
			encoded, err := json.Marshal(pkg)
			require.NoError(t, err)

			require.NotContains(t, string(encoded), `":""`, "empty string value for %s", pkg.ID())
			require.NotContains(t, string(encoded), `":[]`, "empty list value for %s", pkg.ID())
			require.False(t, strings.Contains(string(encoded), `":null`), "null value for %s", pkg.ID())
		}
	})
}
