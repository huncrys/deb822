// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package types

import (
	"strings"

	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/boolean"
	"oaklab.hu/debian/deb822/types/dependency"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/version"
)

// Package represents a Debian package with all its metadata fields.
type Package struct {
	// Name is the name of the binary package.
	Name string `debian:"Package" json:"Package"`
	// Source is the name of the source package from which this package is built.
	Source *dependency.Source `debian:"Source,omitempty" json:"Source,omitzero"`
	// Version is the version of the package.
	Version version.Version `debian:"Version" json:"Version"`
	// InstalledSize is the estimated installed size of the package, in kilobytes.
	InstalledSize *int `debian:"Installed-Size,omitempty" json:"Installed-Size,omitzero"`
	// Maintainer is the name and email address of the person or organization responsible for the package.
	Maintainer string `debian:"Maintainer,omitempty" json:"Maintainer,omitzero"`
	// Architecture is the Debian machine architecture the package is built for.
	Architecture arch.Arch `debian:"Architecture" json:"Architecture"`
	// ArchitectureVariant is an optional field that specifies a variant of the architecture, such as amd64v3 for AMD64 with AVX-512 support.
	// This field is used to distinguish between different variants of the same architecture.
	ArchitectureVariant string `debian:"Architecture-Variant,omitempty" json:"Architecture-Variant,omitzero"`
	// MultiArch is the multi-architecture field, specifying if the package can be installed alongside other architectures.
	// Valid values are "same", "foreign", or the name of an architecture.
	MultiArch string `debian:"Multi-Arch,omitempty" json:"Multi-Arch,omitzero"`
	// Replaces lists other packages that this package replaces.
	Replaces dependency.Dependency `debian:"Replaces,omitempty" json:"Replaces,omitzero"`
	// Breaks lists other packages that this package breaks.
	Breaks dependency.Dependency `debian:"Breaks,omitempty" json:"Breaks,omitzero"`
	// Provides lists virtual packages that this package provides.
	Provides dependency.Dependency `debian:"Provides,omitempty" json:"Provides,omitzero"`
	// Conflicts lists other packages that conflict with this package.
	Conflicts dependency.Dependency `debian:"Conflicts,omitempty" json:"Conflicts,omitzero"`
	// Enhances lists packages that this package enhances.
	Enhances dependency.Dependency `debian:"Enhances,omitempty" json:"Enhances,omitzero"`
	// Depends lists packages that this package depends on.
	Depends dependency.Dependency `debian:"Depends,omitempty" json:"Depends,omitzero"`
	// Recommends lists packages that are recommended to be installed with this package.
	Recommends dependency.Dependency `debian:"Recommends,omitempty" json:"Recommends,omitzero"`
	// Suggests lists packages that are suggested to be installed with this package.
	Suggests dependency.Dependency `debian:"Suggests,omitempty" json:"Suggests,omitzero"`
	// PreDepends lists packages that must be installed and configured before this package.
	PreDepends dependency.Dependency `debian:"Pre-Depends,omitempty" json:"Pre-Depends,omitzero"`
	// Description provides a short description and a long description of the package.
	Description string `debian:"Description,omitempty" json:"Description,omitzero"`
	// Homepage is the URL of the package's homepage, typically where more information can be found.
	Homepage string `debian:"Homepage,omitempty" json:"Homepage,omitzero"`
	// Tag lists tags associated with the package, separated by commas.
	Tag list.CommaDelimited[string] `debian:"Tag,omitempty" json:"Tag,omitzero"`
	// Section categorizes the package within the Debian archive, such as "admin", "devel", or "x11".
	Section string `debian:"Section,omitempty" json:"Section,omitzero"`
	// Priority defines the importance of the package within the Debian system, such as "required", "standard", or "optional".
	Priority string `debian:"Priority,omitempty" json:"Priority,omitzero"`
	// Essential indicates if the package is essential for the system to function. If true, the package cannot be removed.
	Essential *boolean.Boolean `debian:"Essential,omitempty" json:"Essential,omitzero"`
	// Important indicates if the package is important for the system to function. This is less strict than Essential.
	Important *boolean.Boolean `debian:"Important,omitempty" json:"Important,omitzero"`
	// Protected indicates if the package is protected, containing important system boot infrastructure.
	Protected *boolean.Boolean `debian:"Protected,omitempty" json:"Protected,omitzero"`
	// Filename is the name of the package file.
	Filename string `debian:"Filename" json:"Filename"`
	// Size is the size of the package file, in bytes.
	Size int `debian:"Size,omitempty" json:"Size,omitzero"`
	// SHA256 is the SHA-256 checksum of the package file for integrity verification.
	SHA256 string `debian:"SHA256,omitempty" json:"SHA256,omitzero"`
	// DescriptionMD5 is the MD5 checksum of the package description for integrity verification.
	DescriptionMD5 string `debian:"Description-md5,omitempty" json:"Description-md5,omitzero"`
	// MD5sum is the MD5 checksum of the package file for integrity verification.
	MD5sum string `debian:"MD5sum,omitempty" json:"MD5sum,omitzero"`
	// SHA1 is the SHA-1 checksum of the package file for integrity verification.
	SHA1 string `debian:"SHA1,omitempty" json:"SHA1,omitzero"`
	// SHA512 is the SHA-512 checksum of the package file for integrity verification.
	SHA512 string `debian:"SHA512,omitempty" json:"SHA512,omitzero"`

	// Control fields used in the dpkg status file.

	// Status indicates the current status of the package (e.g., "install ok installed").
	Status list.SpaceDelimited[string] `debian:"Status,omitempty" json:"Status,omitzero"`
	// ConfigVersion is the version of the package to which the configuration files belong.
	ConfigVersion *version.Version `debian:"Config-Version,omitempty" json:"Config-Version,omitzero"`
	// Conffiles lists configuration files that are part of the package.
	Conffiles list.NewLineDelimited[string] `debian:"Conffiles,omitempty" json:"Conffiles,omitzero"`
}

// ID returns a unique identifier for the package, combining the name, version, and architecture.
func (p Package) ID() string {
	result := p.Name + "_" + p.Version.String() + "_" + p.Architecture.String()

	if p.ArchitectureVariant != "" {
		result += "_" + p.ArchitectureVariant
	}

	return result
}

// Compare compares two packages by name, version, and architecture.
// It returns an integer comparing the two packages lexicographically.
func (a Package) Compare(b Package) int {
	// Compare package names.
	if cmp := strings.Compare(a.Name, b.Name); cmp != 0 {
		return cmp
	}

	// Compare package versions.
	if cmp := a.Version.Compare(b.Version); cmp != 0 {
		return cmp
	}

	// Compare architectures.
	if a.Architecture.Is(&b.Architecture) || b.Architecture.Is(&a.Architecture) {
		return strings.Compare(a.ArchitectureVariant, b.ArchitectureVariant)
	}

	return strings.Compare(a.Architecture.String(), b.Architecture.String())
}
