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
	Origin string
	// Label provides a human-readable label for the release.
	Label string
	// Suite indicates the suite (such as stable, testing, unstable) the release belongs to.
	Suite string
	// Version denotes the version number of the release.
	Version string
	// Codename is the codename assigned to the release (e.g., "buster", "bullseye").
	Codename string
	// Changelogs provides the URL to the changelogs for the release, detailing changes and updates.
	Changelogs string
	// Date is the timestamp indicating when the release was published.
	Date time.Time
	// ValidUntil specifies the date until which the release is considered valid. It is optional.
	ValidUntil *time.Time `json:"Valid-Until,omitempty"`
	// Architectures lists the CPU architectures supported by the release (e.g., amd64, i386).
	Architectures list.SpaceDelimited[arch.Arch]
	// Components lists the repository components available in the release (e.g., main, contrib, non-free).
	Components list.SpaceDelimited[string]
	// Description provides a brief description of the release.
	Description string
	// MD5Sum lists MD5 checksums for files in the release, used for integrity verification.
	MD5Sum list.NewLineDelimited[filehash.FileHash] `json:",omitempty"`
	// SHA1 lists SHA-1 checksums for files in the release, used for integrity verification.
	SHA1 list.NewLineDelimited[filehash.FileHash] `json:",omitempty"`
	// SHA256 lists SHA-256 checksums for files in the release, used for stronger integrity verification.
	SHA256 list.NewLineDelimited[filehash.FileHash] `json:",omitempty"`
	// AcquireByHash indicates if the release uses hash-based acquisition for file retrieval.
	AcquireByHash *boolean.Boolean `json:"Acquire-By-Hash,omitempty"`
	// SignedBy lists OpenPGP key fingerprints to be used for validating the next Release file.
	SignedBy list.CommaDelimited[string] `json:"Signed-By,omitempty"`
	// https://wiki.debian.org/DebianRepository/Format#No-Support-for-Architecture-all
	NoSupportForArchitectureAll string `json:"No-Support-For-Architecture-all,omitempty"`
	// Snapshots provides the URL to the snapshots for the release.
	Snapshots string
	// NotAutomatic indicates if the package manager should not install packages (or upgrade to newer versions)
	// from this repository without explicit user consent.
	NotAutomatic *boolean.Boolean `json:",omitempty"`
	// ButAutomaticUpgrades indicates if the package manager should automatically install package upgrades from
	// this repository, if the installed version of the package is higher than the version of the package in other
	// sources (APT assigns priority 100).
	ButAutomaticUpgrades *boolean.Boolean `json:",omitempty"`
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
