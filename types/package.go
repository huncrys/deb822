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

// Package represents a Debian package.
type Package struct {
	// Name is the name of the package.
	Name string `json:"Package"`
	// Source is the source package name.
	Source string
	// Version is the version of the package.
	Version version.Version
	// InstalledSize is the installed size of the package, in kilobytes.
	InstalledSize int `json:"Installed-Size,omitempty,string"`
	// Maintainer is the person or organization responsible for the package.
	Maintainer string
	// Architecture is the architecture the package is built for.
	Architecture arch.Arch
	// MultiArch is the multi-architecture hint for the package.
	// Valid values are "same", "foreign", or the name of an architecture.
	MultiArch string `json:"Multi-Arch"`
	// Replaces lists packages that this package replaces.
	Replaces dependency.Dependency
	// Breaks lists packages that this package breaks.
	Breaks dependency.Dependency
	// Provides lists virtual packages that this package provides.
	Provides dependency.Dependency
	// Conflicts lists packages that conflict with this package.
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
	// Description provides a short description of the package.
	Description string
	// Homepage is the URL of the package's homepage.
	Homepage string
	// Tag lists tags associated with the package.
	Tag list.CommaDelimited[string]
	// Section categorizes the package within the archive.
	Section string
	// Priority defines the importance of the package.
	Priority string
	// Essential is true if the package is essential for the system to function.
	Essential *boolean.Boolean `json:",omitempty"`
	// Important is true if the package is important for the system to function.
	// This is a less strict version of Essential.
	Important *boolean.Boolean `json:",omitempty"`
	// Protected is true if the package is protected. Protected packages contain
	// mostly important system boot infrastructure.
	Protected *boolean.Boolean `json:",omitempty"`
	// Filename is the name of the package file.
	Filename string
	// Size is the size of the package file, in bytes.
	Size int `json:",omitempty,string"`
	// MD5sum is the MD5 checksum of the package file.
	MD5sum string
	// SHA256 is the SHA-256 checksum of the package file.
	SHA256 string

	// Control fields used in the dpkg status file.

	// Status is the package status.
	Status list.SpaceDelimited[string] `json:",omitempty"`
	// ConfigVersion is the version of the package to which the configuration
	// files belong.
	ConfigVersion *version.Version `json:"Config-Version,omitempty"`
	// Conffiles lists configuration files that are part of the package.
	Conffiles list.NewLineDelimited[string] `json:",omitempty"`
}

// ID returns a unique identifier for the package.
func (p Package) ID() string {
	return p.Name + "_" + p.Version.String() + "_" + p.Architecture.String()
}

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
	if cmp := strings.Compare(a.Architecture.String(), b.Architecture.String()); cmp != 0 {
		return cmp
	}

	return 0
}
