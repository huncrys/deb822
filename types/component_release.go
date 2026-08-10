// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package types

import (
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/boolean"
)

// ComponentRelease is the per component, per architecture Release stub an
// archive publishes next to the indices it describes, at
// dists/<suite>/<component>/binary-<architecture>/Release. It records which
// slice of which release the directory belongs to; the checksums stay in the
// release wide Release file.
//
// The field order is the one Debian's own stubs are written in.
type ComponentRelease struct {
	// Archive names the suite the indices in this directory belong to (such as stable or unstable).
	Archive string `debian:"Archive" json:"Archive"`
	// Origin specifies the origin of the release, typically indicating the entity that created it.
	Origin string `debian:"Origin,omitempty" json:"Origin,omitzero"`
	// Label provides a human-readable label for the release.
	Label string `debian:"Label,omitempty" json:"Label,omitzero"`
	// Version denotes the version number of the release.
	Version string `debian:"Version,omitempty" json:"Version,omitzero"`
	// AcquireByHash indicates if the release uses hash-based acquisition for file retrieval.
	AcquireByHash *boolean.Boolean `debian:"Acquire-By-Hash,omitempty" json:"Acquire-By-Hash,omitzero"`
	// Component names the repository component the indices belong to (e.g., main, contrib, non-free).
	Component string `debian:"Component" json:"Component"`
	// Architecture is the Debian machine architecture the indices in this directory describe.
	Architecture arch.Arch `debian:"Architecture" json:"Architecture"`
}
