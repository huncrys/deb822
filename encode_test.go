// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Portions of this file are based on code originally from: github.com/paultag/go-debian
 *
 * Copyright (c) Paul R. Tagliamonte <paultag@debian.org>, 2015
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package deb822_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/dpeckett/deb822"
	"github.com/dpeckett/deb822/types/dependency"
	"github.com/dpeckett/deb822/types/version"
	"github.com/stretchr/testify/require"
)

type TestMarshalStruct struct {
	Foo        string
	Version    version.Version
	Dependency dependency.Dependency
}

func TestEncode(t *testing.T) {
	a := TestMarshalStruct{
		Foo:        "Hello",
		Version:    version.MustParse("1.0-1"),
		Dependency: dependency.MustParse("foo, bar (>= 1.0) [amd64] | baz"),
	}

	b := TestMarshalStruct{
		Foo:        "World",
		Version:    version.MustParse("1.0-1"),
		Dependency: dependency.MustParse("foo, bar (>= 2.0) [amd64] | baz"),
	}

	t.Run("Unsigned", func(t *testing.T) {
		var sb strings.Builder
		encoder, err := deb822.NewEncoder(&sb, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, encoder.Close())
		})

		require.NoError(t, encoder.Encode(a))
		require.NoError(t, encoder.Encode(b))

		expected := `Foo: Hello
Version: 1.0-1
Dependency: foo, bar [amd64] (>= 1.0) | baz

Foo: World
Version: 1.0-1
Dependency: foo, bar [amd64] (>= 2.0) | baz
`

		require.Equal(t, expected, sb.String())
	})

	t.Run("Signed", func(t *testing.T) {
		entityConfig := packet.Config{
			RSABits: 1024, // insecure for testing
			Time:    time.Now,
		}

		entity, err := openpgp.NewEntity("test", "", "", &entityConfig)
		require.NoError(t, err)

		var sb strings.Builder
		encoder, err := deb822.NewEncoder(&sb, entity)
		require.NoError(t, err)

		require.NoError(t, encoder.Encode(a))
		require.NoError(t, encoder.Encode(b))

		// Close the encoder to finalize the signature.
		require.NoError(t, encoder.Close())

		signedMessage := sb.String()
		require.Contains(t, signedMessage, "BEGIN PGP SIGNATURE")
		require.Contains(t, signedMessage, "Foo: Hello")
		require.Contains(t, signedMessage, "Foo: World")
		require.Contains(t, signedMessage, "END PGP SIGNATURE")
	})
}
