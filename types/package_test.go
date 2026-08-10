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
	"oaklab.hu/debian/deb822/types/boolean"
	"oaklab.hu/debian/deb822/types/dependency"
	"oaklab.hu/debian/deb822/types/list"
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

// vcardDescription is 2vcard's description as trixie's Translation-en carries
// it, unfolded the way the decoder hands it over: one leading space stripped
// per continuation line, the " ." separator reduced to a blank line, and a
// newline closing every continuation line, the last one included.
const vcardDescription = "convert an addressbook to VCARD file format\n" +
	"2vcard converts address books and alias files into the widely-used\n" +
	"vCard format. Currently it can convert from abook, Eudora, Juno,\n" +
	"LDIF, mutt, mh and pine.\n" +
	"\n" +
	"2vcard was developed using Perl.\n"

// TestDescriptionMD5Sum pins the checksum against values established outside
// this library. The multi-line case is 2vcard from Debian trixie: the archive
// publishes "772b42c5a35b82967966265253189059" as its Description-md5 in both
// main/binary-amd64/Packages and main/i18n/Translation-en, and md5sum(1) over
// the Translation-en field body plus its terminating newline agrees. The single
// line digest was taken the same way, from md5sum(1) over the short
// description and a newline. The empty case is apt's: it records no checksum
// rather than the digest of a bare newline.
func TestDescriptionMD5Sum(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "multi-line with separator",
			description: vcardDescription,
			expected:    "772b42c5a35b82967966265253189059",
		},
		{
			// A description assembled by hand rarely carries the trailing
			// newline the decoder leaves behind. The fold trims it either way,
			// so both spellings have to hash the same.
			name:        "multi-line without trailing newline",
			description: strings.TrimSuffix(vcardDescription, "\n"),
			expected:    "772b42c5a35b82967966265253189059",
		},
		{
			name:        "single line",
			description: "convert an addressbook to VCARD file format",
			expected:    "79a8ae8f714db1af30d480dc5f3a8652",
		},
		{
			name:        "empty",
			description: "",
			expected:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := types.Package{
				Name:         "2vcard",
				Version:      version.MustParse("0.6-4"),
				Architecture: arch.MustParse("all"),
				Description:  test.description,
			}

			require.Equal(t, test.expected, pkg.DescriptionMD5Sum())
		})
	}
}

// TestDescriptionMD5SumFromWire covers the whole path a repository builder
// takes: the description arrives folded, the decoder unfolds it, and the
// checksum still has to come out as the one the archive publishes. Hashing the
// decoded value directly would not.
func TestDescriptionMD5SumFromWire(t *testing.T) {
	stanza := `Package: 2vcard
Version: 0.6-4
Architecture: all
Description: convert an addressbook to VCARD file format
 2vcard converts address books and alias files into the widely-used
 vCard format. Currently it can convert from abook, Eudora, Juno,
 LDIF, mutt, mh and pine.
 .
 2vcard was developed using Perl.
Description-md5: 772b42c5a35b82967966265253189059
`

	decoder, err := deb822.NewDecoder(strings.NewReader(stanza), nil)
	require.NoError(t, err)

	var pkg types.Package
	require.NoError(t, decoder.Decode(&pkg))

	require.Equal(t, vcardDescription, pkg.Description)
	require.Equal(t, pkg.DescriptionMD5, pkg.DescriptionMD5Sum())
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

Package: variant-package
Version: 1.2.3-4
Maintainer: Variant Maintainer <variant@example.com>
Architecture: amd64
Architecture-Variant: amd64v3
Depends: sample-package (>= 1.2)
Description: Sample package for testing
Homepage: https://example.com/variant-package

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

	require.Len(t, packageList, 4)

	rtPackagesBuilder := &strings.Builder{}
	encoder, err := deb822.NewEncoder(rtPackagesBuilder, nil)
	require.NoError(t, err)

	require.NoError(t, encoder.Encode(packageList))

	rtPackages := rtPackagesBuilder.String()
	require.Equal(t, packages, rtPackages)
}

// TestRoundTripPolicyFields covers the Debian Policy and dpkg fields that are
// not exercised by the sample archive data: they have to survive a decode and
// come back out in the very same shape.
func TestRoundTripPolicyFields(t *testing.T) {
	packages := `Package: policy-package
Version: 1.0-1
Architecture: amd64
Built-Using: grub2 (= 2.06-13), libssl (= 3.0.11-1)
Static-Built-Using: golang-1.21 (= 1.21.4-1)
Description: Package exercising the policy fields
Origin: Debian
Bugs: debbugs://bugs.debian.org
Task: desktop, gnome-desktop
Package-Type: udeb
Build-Essential: yes
Subarchitecture: mac
Kernel-Version: 6.1.0-13-amd64
Installer-Menu-Item: 4000
`

	decoder, err := deb822.NewDecoder(strings.NewReader(packages), nil)
	require.NoError(t, err)

	var packageList []types.Package
	require.NoError(t, decoder.Decode(&packageList))

	require.Len(t, packageList, 1)

	pkg := packageList[0]
	require.Equal(t, dependency.MustParse("grub2 (= 2.06-13), libssl (= 3.0.11-1)"), pkg.BuiltUsing)
	require.Equal(t, dependency.MustParse("golang-1.21 (= 1.21.4-1)"), pkg.StaticBuiltUsing)
	require.Equal(t, "Debian", pkg.Origin)
	require.Equal(t, "debbugs://bugs.debian.org", pkg.Bugs)
	require.Equal(t, list.CommaDelimited[string]{"desktop", "gnome-desktop"}, pkg.Task)
	require.Equal(t, "udeb", pkg.PackageType)
	require.NotNil(t, pkg.BuildEssential)
	require.Equal(t, boolean.Boolean(true), *pkg.BuildEssential)
	require.Equal(t, "mac", pkg.Subarchitecture)
	require.Equal(t, "6.1.0-13-amd64", pkg.KernelVersion)
	require.Equal(t, "4000", pkg.InstallerMenuItem)

	rtPackagesBuilder := &strings.Builder{}
	encoder, err := deb822.NewEncoder(rtPackagesBuilder, nil)
	require.NoError(t, err)

	require.NoError(t, encoder.Encode(packageList))

	require.Equal(t, packages, rtPackagesBuilder.String())
}

// TestUnsetPolicyFieldsAreOmitted pins that a package leaving the optional
// fields alone does not grow empty lines for them.
func TestUnsetPolicyFieldsAreOmitted(t *testing.T) {
	pkg := types.Package{
		Name:         "plain-package",
		Version:      version.MustParse("1.0-1"),
		Architecture: arch.MustParse("amd64"),
		Filename:     "pool/main/p/plain-package/plain-package_1.0-1_amd64.deb",
	}

	encodedBuilder := &strings.Builder{}
	encoder, err := deb822.NewEncoder(encodedBuilder, nil)
	require.NoError(t, err)

	require.NoError(t, encoder.Encode([]types.Package{pkg}))

	encoded := encodedBuilder.String()

	for _, key := range []string{
		"Built-Using",
		"Static-Built-Using",
		"Package-Type",
		"Task",
		"Origin",
		"Bugs",
		"Build-Essential",
		"Subarchitecture",
		"Kernel-Version",
		"Installer-Menu-Item",
	} {
		require.NotContains(t, encoded, key+":", "unset %s was written out", key)
	}

	require.Contains(t, encoded, "Package: plain-package\n")
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
		{
			a: types.Package{
				Name:                "pkg",
				Version:             version.MustParse("1.0-1"),
				Architecture:        arch.MustParse("amd64"),
				ArchitectureVariant: "amd64v1",
			},
			b: types.Package{
				Name:                "pkg",
				Version:             version.MustParse("1.0-1"),
				Architecture:        arch.MustParse("amd64"),
				ArchitectureVariant: "amd64v3",
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
				Name:                "pkg",
				Version:             version.MustParse("1.0-1"),
				Architecture:        arch.MustParse("amd64"),
				ArchitectureVariant: "amd64v3",
			},
			expect: -1,
		},
		{
			a: types.Package{
				Name:                "pkg",
				Version:             version.MustParse("1.0-1"),
				Architecture:        arch.MustParse("amd64"),
				ArchitectureVariant: "amd64v3",
			},
			b: types.Package{
				Name:                "pkg",
				Version:             version.MustParse("1.0-1"),
				Architecture:        arch.MustParse("amd64"),
				ArchitectureVariant: "amd64v2",
			},
			expect: 1,
		},
		{
			a: types.Package{
				Name:                "pkg",
				Version:             version.MustParse("1.0-1"),
				Architecture:        arch.MustParse("amd64"),
				ArchitectureVariant: "amd64v3",
			},
			b: types.Package{
				Name:                "pkg",
				Version:             version.MustParse("1.0-1"),
				Architecture:        arch.MustParse("amd64"),
				ArchitectureVariant: "amd64v3",
			},
			expect: 0,
		},
	}

	for _, test := range tests {
		result := test.a.Compare(test.b)
		require.Equal(t, test.expect, result, "Comparing %s and %s", test.a.ID(), test.b.ID())
	}
}
