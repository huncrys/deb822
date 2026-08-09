// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package deb822_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/boolean"
	"oaklab.hu/debian/deb822/types/dependency"
	"oaklab.hu/debian/deb822/types/version"
)

// jsonTagged carries no debian tags at all, the way a consumer's struct did
// before the debian tag existed.
type jsonTagged struct {
	Name       string `json:"Package"`
	MultiArch  string `json:"Multi-Arch,omitempty"`
	Untagged   string
	NeverSeen  string `json:"-"`
	unexported string //nolint:unused // the walker has to skip it
}

// debianTagged puts the debian tag next to a differing json tag, so the two
// can be told apart.
type debianTagged struct {
	Name    string `debian:"Debian-Name" json:"Json-Name"`
	Skipped string `debian:"-" json:"Skipped"`
	Renamed string `debian:"Renamed-Field,omitempty" json:"renamed_field,omitzero"`
}

type inlinedInner struct {
	Inner string `debian:"Inner-Value"`
}

// inlined flattens a named struct field, which is what the inline option is
// for; embedding does the same without it.
type inlined struct {
	Nested inlinedInner `debian:",inline"`
	Value  string       `debian:"Value"`
}

// walkerTypes covers the value shapes the walker has to render and parse
// without any help from encoding/json.
type walkerTypes struct {
	Version   version.Version       `debian:"Version,omitempty"`
	Arch      arch.Arch             `debian:"Architecture,omitempty"`
	Depends   dependency.Dependency `debian:"Depends,omitempty"`
	Size      int                   `debian:"Size,omitempty"`
	Essential boolean.Boolean       `debian:"Essential,omitempty"`
	Optional  *boolean.Boolean      `debian:"Optional,omitempty"`
	Config    *version.Version      `debian:"Config-Version,omitempty"`
	Count     *int                  `debian:"Count,omitempty"`
}

func marshalToString(t *testing.T, data any) string {
	t.Helper()

	var sb strings.Builder
	require.NoError(t, deb822.Marshal(&sb, data))

	return sb.String()
}

func TestTagFallbackToJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		require.Equal(t, `Package: hello
Untagged: there
`, marshalToString(t, jsonTagged{Name: "hello", Untagged: "there", NeverSeen: "nope"}))
	})

	t.Run("unmarshal", func(t *testing.T) {
		var got jsonTagged
		require.NoError(t, deb822.Unmarshal([]byte(`Package: hello
Multi-Arch: same
Untagged: there
NeverSeen: nope
`), &got))

		require.Equal(t, jsonTagged{Name: "hello", MultiArch: "same", Untagged: "there"}, got)
	})

	t.Run("round trip", func(t *testing.T) {
		want := jsonTagged{Name: "hello", MultiArch: "same", Untagged: "there"}

		var got jsonTagged
		require.NoError(t, deb822.Unmarshal([]byte(marshalToString(t, want)), &got))
		require.Equal(t, want, got)
	})
}

func TestDebianTagWins(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		require.Equal(t, `Debian-Name: hello
Renamed-Field: there
`, marshalToString(t, debianTagged{Name: "hello", Skipped: "nope", Renamed: "there"}))
	})

	t.Run("unmarshal", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  debianTagged
		}{
			{
				name:  "debian name matches",
				input: "Debian-Name: hello\n",
				want:  debianTagged{Name: "hello"},
			},
			{
				name:  "json name is not a stanza name",
				input: "Json-Name: hello\n",
				want:  debianTagged{},
			},
			{
				name:  "renamed field matches its debian name",
				input: "Renamed-Field: there\n",
				want:  debianTagged{Renamed: "there"},
			},
			{
				name:  "skipped field is ignored",
				input: "Skipped: nope\n",
				want:  debianTagged{},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				var got debianTagged
				require.NoError(t, deb822.Unmarshal([]byte(test.input), &got))
				require.Equal(t, test.want, got)
			})
		}
	})
}

func TestInlineStruct(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		require.Equal(t, `Inner-Value: deep
Value: shallow
`, marshalToString(t, inlined{Nested: inlinedInner{Inner: "deep"}, Value: "shallow"}))
	})

	t.Run("unmarshal", func(t *testing.T) {
		var got inlined
		require.NoError(t, deb822.Unmarshal([]byte(`Inner-Value: deep
Value: shallow
`), &got))

		require.Equal(t, inlined{Nested: inlinedInner{Inner: "deep"}, Value: "shallow"}, got)
	})
}

// TestEmbeddedStruct pins the flattening of an anonymous field, which is what
// the json:",inline" tag on TestStruct.Fnord relied on under encoding/json.
func TestEmbeddedStruct(t *testing.T) {
	var got TestStruct
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Fnord-Foo-Bar: Thing
`), &got))

	require.Equal(t, "foo", got.Value)
	require.Equal(t, "Thing", got.Fnord.FooBar)

	// The embedded field comes first, at the position it is embedded at, and
	// the untagged boolean renders as "no" because it did not ask for
	// omitempty and "no" is not the empty string.
	require.Equal(t, `Fnord-Foo-Bar: Thing
Value: foo
Extra-Source-Only: no
`, marshalToString(t, got))
}

func TestWalkerValueKinds(t *testing.T) {
	count := 3
	configVersion := version.MustParse("1.0-1")
	essential := boolean.Boolean(true)

	tests := []struct {
		name    string
		value   walkerTypes
		encoded string
	}{
		{
			name:    "zero value renders nothing",
			value:   walkerTypes{},
			encoded: "",
		},
		{
			name: "text marshalers",
			value: walkerTypes{
				Version: version.MustParse("2:1.2.3-4"),
				Arch:    arch.MustParse("amd64"),
				Depends: dependency.MustParse("foo, bar (>= 1.0)"),
			},
			encoded: `Version: 2:1.2.3-4
Architecture: amd64
Depends: foo, bar (>= 1.0)
`,
		},
		{
			name:    "int",
			value:   walkerTypes{Size: 7891488},
			encoded: "Size: 7891488\n",
		},
		{
			name:    "zero int is dropped by omitempty",
			value:   walkerTypes{Size: 0},
			encoded: "",
		},
		{
			name:    "boolean renders as yes",
			value:   walkerTypes{Essential: true},
			encoded: "Essential: yes\n",
		},
		{
			name:    "non nil pointers render their element",
			value:   walkerTypes{Optional: &essential, Config: &configVersion, Count: &count},
			encoded: "Optional: yes\nConfig-Version: 1.0-1\nCount: 3\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.encoded, marshalToString(t, test.value))

			if test.encoded == "" {
				// Nothing was written, so there is no stanza to read back.
				return
			}

			var got walkerTypes
			require.NoError(t, deb822.Unmarshal([]byte(test.encoded), &got))
			require.Equal(t, test.value, got)
		})
	}
}

func TestCaseInsensitiveFieldNames(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "as declared", input: "Debian-Name: foo\n"},
		{name: "lower case", input: "debian-name: foo\n"},
		{name: "upper case", input: "DEBIAN-NAME: foo\n"},
		{name: "mixed case", input: "dEbIaN-nAmE: foo\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got debianTagged
			require.NoError(t, deb822.Unmarshal([]byte(test.input), &got))
			require.Equal(t, "foo", got.Name)
		})
	}
}

func TestEmptyValuesAreDropped(t *testing.T) {
	// A field that renders as the empty string carries no information, so it
	// never makes it into a stanza, whether or not it asked for omitempty.
	require.Equal(t, "Package: hello\n", marshalToString(t, jsonTagged{Name: "hello"}))

	// The same on the way in: a field present with an empty value is treated
	// as absent, so a parsed type is never handed an empty string.
	var got walkerTypes
	require.NoError(t, deb822.Unmarshal([]byte("Version:\nSize: 4\n"), &got))
	require.Equal(t, walkerTypes{Size: 4}, got)
}
