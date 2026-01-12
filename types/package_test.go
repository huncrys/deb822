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
	"compress/gzip"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/dependency"
	"oaklab.hu/debian/deb822/types/version"
)

func TestPackage(t *testing.T) {
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

	var packageList []types.Package
	require.NoError(t, decoder.Decode(&packageList))

	require.Len(t, packageList, 1324)

	expectedInstalledSize := 28591

	t.Run("package", func(t *testing.T) {
		expected := types.Package{
			Name:           "0ad",
			Version:        version.MustParse("0.0.26-3"),
			InstalledSize:  &expectedInstalledSize,
			Maintainer:     "Debian Games Team <pkg-games-devel@lists.alioth.debian.org>",
			Architecture:   arch.MustParse("amd64"),
			Depends:        dependency.MustParse("0ad-data (>= 0.0.26), 0ad-data (<= 0.0.26-3), 0ad-data-common (>= 0.0.26), 0ad-data-common (<= 0.0.26-3), libboost-filesystem1.74.0 (>= 1.74.0), libc6 (>= 2.34), libcurl3-gnutls (>= 7.32.0), libenet7, libfmt9 (>= 9.1.0+ds1), libfreetype6 (>= 2.2.1), libgcc-s1 (>= 3.4), libgloox18 (>= 1.0.24), libicu72 (>= 72.1~rc-1~), libminiupnpc17 (>= 1.9.20140610), libopenal1 (>= 1.14), libpng16-16 (>= 1.6.2-1), libsdl2-2.0-0 (>= 2.0.12), libsodium23 (>= 1.0.14), libstdc++6 (>= 12), libvorbisfile3 (>= 1.1.2), libwxbase3.2-1 (>= 3.2.1+dfsg), libwxgtk-gl3.2-1 (>= 3.2.1+dfsg), libwxgtk3.2-1 (>= 3.2.1+dfsg-2), libx11-6, libxml2 (>= 2.9.0), zlib1g (>= 1:1.2.0)"),
			PreDepends:     dependency.MustParse("dpkg (>= 1.15.6~)"),
			Description:    "Real-time strategy game of ancient warfare",
			Homepage:       "https://play0ad.com/",
			Tag:            []string{"game::strategy", "interface::graphical", "interface::x11", "role::program", "uitoolkit::sdl", "uitoolkit::wxwidgets", "use::gameplaying", "x11::application"},
			Section:        "games",
			Priority:       "optional",
			Filename:       "pool/main/0/0ad/0ad_0.0.26-3_amd64.deb",
			Size:           7891488,
			SHA256:         "3a2118df47bf3f04285649f0455c2fc6fe2dc7f0b237073038aa00af41f0d5f2",
			DescriptionMD5: "d943033bedada21853d2ae54a2578a7b",
			MD5sum:         "4d471183a39a3a11d00cd35bf9f6803d",
		}

		require.Equal(t, expected, packageList[0])
	})

	t.Run("source", func(t *testing.T) {
		expectedVersion := version.MustParse("0.1.6-2")
		expected := dependency.Source{
			Name:    "2048-qt",
			Version: &expectedVersion,
		}

		require.Equal(t, &expected, packageList[5].Source)
	})

	t.Run("ID", func(t *testing.T) {
		require.Equal(t, "0ad_0.0.26-3_amd64", packageList[0].ID())
		require.Equal(t, "2048-qt_0.1.6-2+b2_amd64", packageList[5].ID())
	})
}

func TestRoundTrip(t *testing.T) {
	packages := `Package: sample-package
Version: 1.2.3-4
Maintainer: Sample Maintainer <sample@example.com>
Architecture: amd64
Depends: libsample1 (>= 1.0), libsample2
Description: Sample package for testing
 A longer description of the sample package.
Homepage: https://example.com/sample-package
SHA256: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
Description-md5: d41d8cd98f00b204e9800998ecf8427e
MD5sum: d41d8cd98f00b204e9800998ecf8427e

Package: another-package
Source: source-package (1.0.0-1)
Version: 0.9.8-1
Maintainer: Another Maintainer <another@example.com>
Architecture: all
Depends: sample-package (>= 1.2)
Description: Another sample package
Homepage: https://example.com/another-package

Package: another-package
Source: source-package
Version: 0.9.8
Maintainer: Another Maintainer <another@example.com>
Architecture: all
Depends: sample-package (>= 1.2)
Description: Another sample package without source version
Homepage: https://example.com/another-package
`

	decoder, err := deb822.NewDecoder(strings.NewReader(packages), nil)
	require.NoError(t, err)

	var packageList []types.Package
	require.NoError(t, decoder.Decode(&packageList))

	require.Len(t, packageList, 3)

	rtPackagesBuilder := &strings.Builder{}
	encoder, err := deb822.NewEncoder(rtPackagesBuilder, nil)
	require.NoError(t, err)

	require.NoError(t, encoder.Encode(packageList))

	rtPackages := rtPackagesBuilder.String()
	require.Equal(t, packages, rtPackages)
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b   types.Package
		expect int
	}{
		{
			a: types.Package{
				Name:    "pkg",
				Version: version.MustParse("1.0-1"),
			},
			b: types.Package{
				Name:    "pkg",
				Version: version.MustParse("1.0-2"),
			},
			expect: -1,
		},
		{
			a: types.Package{
				Name:    "pkg",
				Version: version.MustParse("2.0-1"),
			},
			b: types.Package{
				Name:    "pkg",
				Version: version.MustParse("1.9-9"),
			},
			expect: 1,
		},
		{
			a: types.Package{
				Name:    "pkg",
				Version: version.MustParse("1.0-1"),
			},
			b: types.Package{
				Name:    "pkg",
				Version: version.MustParse("1.0-1"),
			},
			expect: 0,
		},
		{
			a: types.Package{
				Name:    "pkgA",
				Version: version.MustParse("1.0-1"),
			},
			b: types.Package{
				Name:    "pkgB",
				Version: version.MustParse("1.0-1"),
			},
			expect: -1,
		},
		{
			a: types.Package{
				Name:         "pkg",
				Version:      version.MustParse("1.0-1"),
				Architecture: arch.MustParse("amd64"),
			},
			b: types.Package{
				Name:         "pkg",
				Version:      version.MustParse("1.0-1"),
				Architecture: arch.MustParse("arm64"),
			},
			expect: -1,
		},
	}

	for _, test := range tests {
		result := test.a.Compare(test.b)
		require.Equal(t, test.expect, result, "Comparing %s and %s", test.a.ID(), test.b.ID())
	}
}
