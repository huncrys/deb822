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
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/time"
	"oaklab.hu/debian/deb822/types/version"
)

// Changes is a Debian changes file, as described by deb-changes(5). It is the
// single stanza of a .changes file, which is usually OpenPGP clearsigned, and
// it describes one upload: which files it carries and what changed in it.
type Changes struct {
	// Format is the changes file format version, such as "1.8".
	Format string `debian:"Format" json:"Format"`
	// Date is the date the source package was built, taken from the top entry of the changelog.
	Date time.Time `debian:"Date" json:"Date"`
	// Source is the name of the source package, optionally followed by the source
	// version in parentheses when it differs from the binary version.
	Source string `debian:"Source" json:"Source"`
	// Binary lists the binary packages the upload carries. Note the separator:
	// unlike the comma separated Binary field of a .dsc, the .changes one is
	// space separated.
	Binary list.SpaceDelimited[string] `debian:"Binary,omitempty" json:"Binary,omitzero"`
	// Architecture lists the architectures the upload carries files for.
	//
	// Besides real architectures the field also uses the pseudo architecture
	// "source", which marks an upload that includes the source package. It is
	// not special cased here: as a single unknown token it parses into the
	// ordinary tuple base-gnu-linux-source, and renders back as "source".
	Architecture list.SpaceDelimited[arch.Arch] `debian:"Architecture,omitempty" json:"Architecture,omitzero"`
	// Version is the version of the package, with its epoch if it has one.
	Version version.Version `debian:"Version" json:"Version"`
	// Distribution lists the distributions the package is to be installed into.
	Distribution list.SpaceDelimited[string] `debian:"Distribution" json:"Distribution"`
	// Urgency is the urgency of the upload, such as "low", "medium" or "high".
	Urgency string `debian:"Urgency" json:"Urgency"`
	// Maintainer is the name and email address of the person or organization responsible for the package.
	Maintainer string `debian:"Maintainer" json:"Maintainer"`
	// ChangedBy is the name and email address of the person who prepared this
	// upload, which need not be the maintainer.
	ChangedBy string `debian:"Changed-By,omitempty" json:"Changed-By,omitzero"`
	// Description holds one line per binary package in the upload, each naming
	// the package and its short description.
	Description string `debian:"Description" json:"Description"`
	// Closes lists the bug numbers this upload closes.
	Closes list.SpaceDelimited[string] `debian:"Closes,omitempty" json:"Closes,omitzero"`
	// Changes is the changelog entries of this upload, verbatim, including the
	// blank lines that a deb822 stanza carries as a lone "." on a line.
	Changes string `debian:"Changes" json:"Changes"`
	// ChecksumsSha1 lists the files of the upload with their SHA-1 checksums.
	ChecksumsSha1 list.NewLineDelimited[filehash.FileHash] `debian:"Checksums-Sha1,omitempty" json:"Checksums-Sha1,omitzero"`
	// ChecksumsSha256 lists the files of the upload with their SHA-256 checksums.
	ChecksumsSha256 list.NewLineDelimited[filehash.FileHash] `debian:"Checksums-Sha256,omitempty" json:"Checksums-Sha256,omitzero"`
	// Files lists the files of the upload with their MD5 checksums, archive
	// section and priority.
	Files list.NewLineDelimited[filehash.ChangesFileHash] `debian:"Files" json:"Files"`
}
