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

package arch_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822/types/arch"
)

func TestArchBasics(t *testing.T) {
	a, err := arch.Parse("amd64")
	require.NoError(t, err)

	require.Equal(t, "amd64", a.CPU)
	require.Equal(t, "gnu", a.ABI)
	require.Equal(t, "linux", a.OS)
}

func TestArchCompareBasics(t *testing.T) {
	a, err := arch.Parse("amd64")
	require.NoError(t, err)

	equivs := []string{
		"gnu-linux-amd64",
		"linux-amd64",
		"linux-any",
		"any",
		"gnu-linux-any",
	}

	for _, el := range equivs {
		other, err := arch.Parse(el)
		require.NoError(t, err)

		require.True(t, a.Is(&other))
		require.True(t, other.Is(&a))
	}

	unequivs := []string{
		"gnu-linux-all",
		"all",

		"gnuu-linux-amd64",
		"gnu-linuxx-amd64",
		"gnu-linux-amd644",
	}

	for _, el := range unequivs {
		other, err := arch.Parse(el)
		require.NoError(t, err)

		require.False(t, a.Is(&other))
		require.False(t, other.Is(&a))
	}
}

func TestArchCompareAllAny(t *testing.T) {
	all, err := arch.Parse("all")
	require.NoError(t, err)

	wildcard, err := arch.Parse("any")
	require.NoError(t, err)

	require.False(t, all.Is(&wildcard))
	require.False(t, wildcard.Is(&all))
	require.False(t, wildcard.Is(&wildcard))
}

func TestMustParse(t *testing.T) {
	a := arch.MustParse("amd64")

	require.Equal(t, "amd64", a.CPU)
	require.Equal(t, "gnu", a.ABI)
	require.Equal(t, "linux", a.OS)

	require.Panics(t, func() {
		arch.MustParse("a-b-c-d")
	})
}
func TestArchString(t *testing.T) {
	tests := []struct {
		name     string
		arch     arch.Arch
		expected string
	}{
		{
			name:     "standard amd64",
			arch:     arch.Arch{ABI: "gnu", OS: "linux", CPU: "amd64"},
			expected: "amd64",
		},
		{
			name:     "non-standard ABI",
			arch:     arch.Arch{ABI: "musl", OS: "linux", CPU: "amd64"},
			expected: "musl-amd64",
		},
		{
			name:     "non-standard OS",
			arch:     arch.Arch{ABI: "gnu", OS: "kfreebsd", CPU: "amd64"},
			expected: "kfreebsd-amd64",
		},
		{
			name:     "non-standard ABI and OS",
			arch:     arch.Arch{ABI: "bsd", OS: "openbsd", CPU: "i386"},
			expected: "bsd-openbsd-i386",
		},
		{
			name:     "any wildcard",
			arch:     arch.Arch{ABI: "any", OS: "any", CPU: "any"},
			expected: "any",
		},
		{
			name:     "all",
			arch:     arch.Arch{ABI: "all", OS: "all", CPU: "all"},
			expected: "all",
		},
		{
			name:     "OS wildcard",
			arch:     arch.Arch{ABI: "gnu", OS: "any", CPU: "amd64"},
			expected: "amd64",
		},
		{
			name:     "CPU only",
			arch:     arch.Arch{ABI: "any", OS: "any", CPU: "arm64"},
			expected: "arm64",
		},
		{
			name:     "empty ABI with non-standard OS",
			arch:     arch.Arch{ABI: "", OS: "kfreebsd", CPU: "amd64"},
			expected: "kfreebsd-amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arch.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestArchMarshalText(t *testing.T) {
	a := arch.Arch{ABI: "gnu", OS: "linux", CPU: "amd64"}

	text, err := a.MarshalText()
	require.NoError(t, err)

	require.Equal(t, "amd64", string(text))
}

func TestArchUnmarshalText(t *testing.T) {
	var a arch.Arch

	err := a.UnmarshalText([]byte("amd64"))
	require.NoError(t, err)

	require.Equal(t, arch.Arch{ABI: "gnu", OS: "linux", CPU: "amd64"}, a)
}
