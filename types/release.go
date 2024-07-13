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

	"github.com/dpeckett/deb822/types/arch"
	"github.com/dpeckett/deb822/types/boolean"
	"github.com/dpeckett/deb822/types/filehash"
	"github.com/dpeckett/deb822/types/list"
	"github.com/dpeckett/deb822/types/time"
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
	// SHA256 lists SHA-256 checksums for files in the release, used for stronger integrity verification.
	SHA256 list.NewLineDelimited[filehash.FileHash]
	// AcquireByHash indicates if the release uses hash-based acquisition for file retrieval.
	AcquireByHash *boolean.Boolean `json:"Acquire-By-Hash,omitempty"`
	// SignedBy lists OpenPGP key fingerprints to be used for validating the next Release file.
	SignedBy list.CommaDelimited[string] `json:"Signed-By,omitempty"`
}

// SHA256Sums returns a map of SHA-256 checksums for files in the release.
func (r *Release) SHA256Sums() (map[string][]byte, error) {
	ret := make(map[string][]byte)
	for _, hash := range r.SHA256 {
		var err error
		ret[hash.Filename], err = hex.DecodeString(hash.Hash)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}
