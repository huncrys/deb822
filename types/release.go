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
	"encoding/hex"

	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/boolean"
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/time"
)

// Release represents a Debian release with its associated metadata.
type Release struct {
	// Origin specifies the origin of the release, typically indicating the entity that created it.
	Origin string `debian:"Origin" json:"Origin"`
	// Label provides a human-readable label for the release.
	Label string `debian:"Label,omitempty" json:"Label,omitzero"`
	// Suite indicates the suite (such as stable, testing, unstable) the release belongs to.
	Suite string `debian:"Suite" json:"Suite"`
	// Version denotes the version number of the release.
	Version string `debian:"Version,omitempty" json:"Version,omitzero"`
	// Codename is the codename assigned to the release (e.g., "buster", "bullseye").
	Codename string `debian:"Codename" json:"Codename"`
	// Changelogs provides the URL to the changelogs for the release, detailing changes and updates.
	Changelogs string `debian:"Changelogs,omitempty" json:"Changelogs,omitzero"`
	// Date is the timestamp indicating when the release was published.
	Date time.Time `debian:"Date" json:"Date"`
	// ValidUntil specifies the date until which the release is considered valid. It is optional.
	ValidUntil *time.Time `debian:"Valid-Until,omitempty" json:"Valid-Until,omitzero"`
	// Architectures lists the CPU architectures supported by the release (e.g., amd64, i386).
	Architectures list.SpaceDelimited[arch.Arch] `debian:"Architectures" json:"Architectures"`
	// Components lists the repository components available in the release (e.g., main, contrib, non-free).
	Components list.SpaceDelimited[string] `debian:"Components" json:"Components"`
	// Description provides a brief description of the release.
	Description string `debian:"Description,omitempty" json:"Description,omitzero"`
	// MD5Sum lists MD5 checksums for files in the release, used for integrity verification.
	MD5Sum list.NewLineDelimited[filehash.FileHash] `debian:"MD5Sum,omitempty" json:"MD5Sum,omitzero"`
	// SHA1 lists SHA-1 checksums for files in the release, used for integrity verification.
	SHA1 list.NewLineDelimited[filehash.FileHash] `debian:"SHA1,omitempty" json:"SHA1,omitzero"`
	// SHA256 lists SHA-256 checksums for files in the release, used for stronger integrity verification.
	SHA256 list.NewLineDelimited[filehash.FileHash] `debian:"SHA256,omitempty" json:"SHA256,omitzero"`
	// SHA512 lists SHA-512 checksums for files in the release, used for stronger integrity verification.
	SHA512 list.NewLineDelimited[filehash.FileHash] `debian:"SHA512,omitempty" json:"SHA512,omitzero"`
	// AcquireByHash indicates if the release uses hash-based acquisition for file retrieval.
	AcquireByHash *boolean.Boolean `debian:"Acquire-By-Hash,omitempty" json:"Acquire-By-Hash,omitzero"`
	// SignedBy lists OpenPGP key fingerprints to be used for validating the next Release file.
	SignedBy list.CommaDelimited[string] `debian:"Signed-By,omitempty" json:"Signed-By,omitzero"`
	// https://wiki.debian.org/DebianRepository/Format#No-Support-for-Architecture-all
	NoSupportForArchitectureAll string `debian:"No-Support-for-Architecture-all,omitempty" json:"No-Support-for-Architecture-all,omitzero"`
	// Snapshots provides the URL to the snapshots for the release.
	Snapshots string `debian:"Snapshots,omitempty" json:"Snapshots,omitzero"`
	// NotAutomatic indicates if the package manager should not install packages (or upgrade to newer versions)
	// from this repository without explicit user consent.
	NotAutomatic *boolean.Boolean `debian:"NotAutomatic,omitempty" json:"NotAutomatic,omitzero"`
	// ButAutomaticUpgrades indicates if the package manager should automatically install package upgrades from
	// this repository, if the installed version of the package is higher than the version of the package in other
	// sources (APT assigns priority 100).
	ButAutomaticUpgrades *boolean.Boolean `debian:"ButAutomaticUpgrades,omitempty" json:"ButAutomaticUpgrades,omitzero"`
}

func sums(hashes list.NewLineDelimited[filehash.FileHash]) (map[string][]byte, error) {
	ret := make(map[string][]byte)
	for _, hash := range hashes {
		var err error
		ret[hash.Filename], err = hex.DecodeString(hash.Hash)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// MD5Sums returns a map of MD5 checksums for files in the release.
func (r *Release) MD5Sums() (map[string][]byte, error) {
	return sums(r.MD5Sum)
}

// SHA1Sums returns a map of SHA-1 checksums for files in the release.
func (r *Release) SHA1Sums() (map[string][]byte, error) {
	return sums(r.SHA1)
}

// SHA256Sums returns a map of SHA-256 checksums for files in the release.
func (r *Release) SHA256Sums() (map[string][]byte, error) {
	return sums(r.SHA256)
}

// SHA512Sums returns a map of SHA-512 checksums for files in the release.
func (r *Release) SHA512Sums() (map[string][]byte, error) {
	return sums(r.SHA512)
}
