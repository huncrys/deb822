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
	require.Equal(t, "base", a.ABI)
	require.Equal(t, "gnu", a.Libc)
	require.Equal(t, "linux", a.OS)
}

func TestArchParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected arch.Arch
	}{
		{
			name:     "single component",
			input:    "amd64",
			expected: arch.Arch{ABI: "base", Libc: "gnu", OS: "linux", CPU: "amd64"},
		},
		{
			name:     "all",
			input:    "all",
			expected: arch.Arch{ABI: "all", Libc: "all", OS: "all", CPU: "all"},
		},
		{
			name:     "two components are os-cpu",
			input:    "kfreebsd-amd64",
			expected: arch.Arch{ABI: "base", Libc: "gnu", OS: "kfreebsd", CPU: "amd64"},
		},
		{
			name:     "three components are libc-os-cpu",
			input:    "musl-linux-amd64",
			expected: arch.Arch{ABI: "base", Libc: "musl", OS: "linux", CPU: "amd64"},
		},
		{
			name:     "four components are explicit",
			input:    "eabihf-gnu-linux-armhf",
			expected: arch.Arch{ABI: "eabihf", Libc: "gnu", OS: "linux", CPU: "armhf"},
		},
		{
			name:     "bare any wildcard",
			input:    "any",
			expected: arch.Arch{ABI: "any", Libc: "any", OS: "any", CPU: "any"},
		},
		{
			name:     "wildcard is left padded",
			input:    "any-amd64",
			expected: arch.Arch{ABI: "any", Libc: "any", OS: "any", CPU: "amd64"},
		},
		{
			name:     "os wildcard is left padded",
			input:    "linux-any",
			expected: arch.Arch{ABI: "any", Libc: "any", OS: "linux", CPU: "any"},
		},
		{
			name:     "libc wildcard is left padded",
			input:    "musl-linux-any",
			expected: arch.Arch{ABI: "any", Libc: "musl", OS: "linux", CPU: "any"},
		},
		{
			name:     "four component wildcard is not padded",
			input:    "any-any-linux-any",
			expected: arch.Arch{ABI: "any", Libc: "any", OS: "linux", CPU: "any"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := arch.Parse(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, a)

			/* Parse and UnmarshalText must agree on every path. */
			var b arch.Arch
			require.NoError(t, b.UnmarshalText([]byte(tt.input)))
			require.Equal(t, tt.expected, b)
		})
	}
}

func TestArchParseInvalid(t *testing.T) {
	for _, input := range []string{"", "a-b-c-d-e"} {
		t.Run(input, func(t *testing.T) {
			_, err := arch.Parse(input)
			require.Error(t, err)
		})
	}
}

func TestArchRoundTrip(t *testing.T) {
	inputs := []string{
		"all",
		"amd64",
		"i386",
		"kfreebsd-amd64",
		"musl-linux-amd64",
		"eabihf-gnu-linux-armhf",
		"any",
		"any-amd64",
		"linux-any",
		"musl-linux-any",
		"any-any-linux-any",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			a := arch.MustParse(input)
			require.Equal(t, a, arch.MustParse(a.String()))
		})
	}
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

func TestArchMatching(t *testing.T) {
	tests := []struct {
		name     string
		arch     arch.Arch
		other    arch.Arch
		expected bool
	}{
		{
			name:     "cpu wildcard matches",
			arch:     arch.MustParse("any-amd64"),
			other:    arch.MustParse("amd64"),
			expected: true,
		},
		{
			name:     "cpu wildcard does not match another cpu",
			arch:     arch.MustParse("any-amd64"),
			other:    arch.MustParse("arm64"),
			expected: false,
		},
		{
			name:     "libc wildcard matches",
			arch:     arch.MustParse("musl-linux-any"),
			other:    arch.MustParse("musl-linux-amd64"),
			expected: true,
		},
		{
			name:     "libc wildcard does not match another libc",
			arch:     arch.MustParse("musl-linux-any"),
			other:    arch.MustParse("amd64"),
			expected: false,
		},
		{
			name:     "any does not match all",
			arch:     arch.MustParse("any"),
			other:    arch.MustParse("all"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.arch.Is(&tt.other))
			require.Equal(t, tt.expected, tt.other.Is(&tt.arch))
		})
	}
}

func TestMustParse(t *testing.T) {
	a := arch.MustParse("amd64")

	require.Equal(t, "amd64", a.CPU)
	require.Equal(t, "base", a.ABI)
	require.Equal(t, "gnu", a.Libc)
	require.Equal(t, "linux", a.OS)

	require.Panics(t, func() {
		arch.MustParse("a-b-c-d-e")
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
			arch:     arch.Arch{ABI: "base", Libc: "gnu", OS: "linux", CPU: "amd64"},
			expected: "amd64",
		},
		{
			name:     "non-standard libc",
			arch:     arch.Arch{ABI: "base", Libc: "musl", OS: "linux", CPU: "amd64"},
			expected: "musl-linux-amd64",
		},
		{
			name:     "non-standard OS",
			arch:     arch.Arch{ABI: "base", Libc: "gnu", OS: "kfreebsd", CPU: "amd64"},
			expected: "kfreebsd-amd64",
		},
		{
			name:     "non-standard libc and OS",
			arch:     arch.Arch{ABI: "base", Libc: "bsd", OS: "openbsd", CPU: "i386"},
			expected: "bsd-openbsd-i386",
		},
		{
			name:     "non-standard ABI",
			arch:     arch.Arch{ABI: "eabihf", Libc: "gnu", OS: "linux", CPU: "armhf"},
			expected: "eabihf-gnu-linux-armhf",
		},
		{
			name:     "any wildcard",
			arch:     arch.Arch{ABI: "any", Libc: "any", OS: "any", CPU: "any"},
			expected: "any",
		},
		{
			name:     "all",
			arch:     arch.Arch{ABI: "all", Libc: "all", OS: "all", CPU: "all"},
			expected: "all",
		},
		{
			name:     "CPU wildcard",
			arch:     arch.Arch{ABI: "any", Libc: "any", OS: "any", CPU: "arm64"},
			expected: "any-arm64",
		},
		{
			name:     "OS wildcard",
			arch:     arch.Arch{ABI: "any", Libc: "any", OS: "linux", CPU: "any"},
			expected: "linux-any",
		},
		{
			name:     "libc wildcard",
			arch:     arch.Arch{ABI: "any", Libc: "musl", OS: "linux", CPU: "any"},
			expected: "musl-linux-any",
		},
		{
			name:     "zero value",
			arch:     arch.Arch{},
			expected: "",
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
	a := arch.Arch{ABI: "base", Libc: "gnu", OS: "linux", CPU: "amd64"}

	text, err := a.MarshalText()
	require.NoError(t, err)

	require.Equal(t, "amd64", string(text))

	text, err = arch.Arch{}.MarshalText()
	require.NoError(t, err)

	require.Empty(t, string(text))
}

func TestArchUnmarshalText(t *testing.T) {
	var a arch.Arch

	err := a.UnmarshalText([]byte("amd64"))
	require.NoError(t, err)

	require.Equal(t, arch.Arch{ABI: "base", Libc: "gnu", OS: "linux", CPU: "amd64"}, a)
}
