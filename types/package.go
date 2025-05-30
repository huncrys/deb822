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

	"github.com/dpeckett/deb822/types/arch"
	"github.com/dpeckett/deb822/types/boolean"
	"github.com/dpeckett/deb822/types/dependency"
	"github.com/dpeckett/deb822/types/list"
	"github.com/dpeckett/deb822/types/version"
)

// Package represents a Debian package with all its metadata fields.
type Package struct {
	// Name is the name of the binary package.
	Name string `json:"Package"`
	// Source is the name of the source package from which this package is built.
	Source string
	// Version is the version of the package.
	Version version.Version
	// InstalledSize is the estimated installed size of the package, in kilobytes.
	InstalledSize *int `json:"Installed-Size,omitempty,string"`
	// Maintainer is the name and email address of the person or organization responsible for the package.
	Maintainer string
	// Architecture is the Debian machine architecture the package is built for.
	Architecture arch.Arch
	// MultiArch is the multi-architecture field, specifying if the package can be installed alongside other architectures.
	// Valid values are "same", "foreign", or the name of an architecture.
	MultiArch string `json:"Multi-Arch"`
	// Replaces lists other packages that this package replaces.
	Replaces dependency.Dependency
	// Breaks lists other packages that this package breaks.
	Breaks dependency.Dependency
	// Provides lists virtual packages that this package provides.
	Provides dependency.Dependency
	// Conflicts lists other packages that conflict with this package.
	Conflicts dependency.Dependency
	// Enhances lists packages that this package enhances.
	Enhances dependency.Dependency
	// Depends lists packages that this package depends on.
	Depends dependency.Dependency
	// Recommends lists packages that are recommended to be installed with this package.
	Recommends dependency.Dependency
	// Suggests lists packages that are suggested to be installed with this package.
	Suggests dependency.Dependency
	// PreDepends lists packages that must be installed and configured before this package.
	PreDepends dependency.Dependency `json:"Pre-Depends"`
	// Description provides a short description and a long description of the package.
	Description string
	// Homepage is the URL of the package's homepage, typically where more information can be found.
	Homepage string
	// Tag lists tags associated with the package, separated by commas.
	Tag list.CommaDelimited[string]
	// Section categorizes the package within the Debian archive, such as "admin", "devel", or "x11".
	Section string
	// Priority defines the importance of the package within the Debian system, such as "required", "standard", or "optional".
	Priority string
	// Essential indicates if the package is essential for the system to function. If true, the package cannot be removed.
	Essential *boolean.Boolean `json:",omitempty"`
	// Important indicates if the package is important for the system to function. This is less strict than Essential.
	Important *boolean.Boolean `json:",omitempty"`
	// Protected indicates if the package is protected, containing important system boot infrastructure.
	Protected *boolean.Boolean `json:",omitempty"`
	// Filename is the name of the package file.
	Filename string
	// Size is the size of the package file, in bytes.
	Size int `json:",omitempty,string"`
	// SHA256 is the SHA-256 checksum of the package file for integrity verification.
	SHA256 string

	// Control fields used in the dpkg status file.

	// Status indicates the current status of the package (e.g., "install ok installed").
	Status list.SpaceDelimited[string] `json:",omitempty"`
	// ConfigVersion is the version of the package to which the configuration files belong.
	ConfigVersion *version.Version `json:"Config-Version,omitempty"`
	// Conffiles lists configuration files that are part of the package.
	Conffiles list.NewLineDelimited[string] `json:",omitempty"`
}

// ID returns a unique identifier for the package, combining the name, version, and architecture.
func (p Package) ID() string {
	return p.Name + "_" + p.Version.String() + "_" + p.Architecture.String()
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
		return 0
	}

	return strings.Compare(a.Architecture.String(), b.Architecture.String())
}
