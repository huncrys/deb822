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
	"oaklab.hu/debian/deb822/types/dependency"
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/version"
)

// Dsc is a Debian source control file, as described by dsc(5). It is the single
// stanza of a .dsc file, which is usually OpenPGP clearsigned.
//
// A Dsc describes a source package as it was built by the packaging tools. The
// archive publishes the same information, plus its own pool bookkeeping, as a
// Source stanza in a Sources index.
type Dsc struct {
	// Format is the source package format, such as "3.0 (quilt)".
	Format string `debian:"Format" json:"Format"`
	// Source is the name of the source package.
	Source string `debian:"Source" json:"Source"`
	// Binary lists the binary packages this source package builds.
	Binary list.CommaDelimited[string] `debian:"Binary,omitempty" json:"Binary,omitzero"`
	// Architecture lists the architectures the source package can be built for.
	Architecture list.SpaceDelimited[arch.Arch] `debian:"Architecture,omitempty" json:"Architecture,omitzero"`
	// Version is the version of the source package.
	Version version.Version `debian:"Version" json:"Version"`
	// Origin is the distribution the package originally came from.
	Origin string `debian:"Origin,omitempty" json:"Origin,omitzero"`
	// Maintainer is the name and email address of the person or organization responsible for the package.
	Maintainer string `debian:"Maintainer" json:"Maintainer"`
	// Uploaders lists co-maintainers allowed to upload the package.
	Uploaders list.CommaDelimited[string] `debian:"Uploaders,omitempty" json:"Uploaders,omitzero"`
	// Homepage is the URL of the upstream project's homepage.
	Homepage string `debian:"Homepage,omitempty" json:"Homepage,omitzero"`
	// StandardsVersion is the version of the Debian Policy the package claims to comply with.
	StandardsVersion string `debian:"Standards-Version,omitempty" json:"Standards-Version,omitzero"`
	// VcsBrowser is a URL to a web interface browsing the packaging repository.
	VcsBrowser string `debian:"Vcs-Browser,omitempty" json:"Vcs-Browser,omitzero"`
	// VcsArch is the location of the packaging repository, in GNU arch.
	VcsArch string `debian:"Vcs-Arch,omitempty" json:"Vcs-Arch,omitzero"`
	// VcsBzr is the location of the packaging repository, in Bazaar.
	VcsBzr string `debian:"Vcs-Bzr,omitempty" json:"Vcs-Bzr,omitzero"`
	// VcsCvs is the location of the packaging repository, in CVS.
	VcsCvs string `debian:"Vcs-Cvs,omitempty" json:"Vcs-Cvs,omitzero"`
	// VcsDarcs is the location of the packaging repository, in Darcs.
	VcsDarcs string `debian:"Vcs-Darcs,omitempty" json:"Vcs-Darcs,omitzero"`
	// VcsGit is the location of the packaging repository, in Git.
	VcsGit string `debian:"Vcs-Git,omitempty" json:"Vcs-Git,omitzero"`
	// VcsHg is the location of the packaging repository, in Mercurial.
	VcsHg string `debian:"Vcs-Hg,omitempty" json:"Vcs-Hg,omitzero"`
	// VcsMtn is the location of the packaging repository, in Monotone.
	VcsMtn string `debian:"Vcs-Mtn,omitempty" json:"Vcs-Mtn,omitzero"`
	// VcsSvn is the location of the packaging repository, in Subversion.
	VcsSvn string `debian:"Vcs-Svn,omitempty" json:"Vcs-Svn,omitzero"`
	// Testsuite names the automatic test suites the package ships, such as "autopkgtest".
	Testsuite string `debian:"Testsuite,omitempty" json:"Testsuite,omitzero"`
	// TestsuiteTriggers lists the binary packages whose upload should trigger a run of the test suite.
	TestsuiteTriggers string `debian:"Testsuite-Triggers,omitempty" json:"Testsuite-Triggers,omitzero"`
	// BuildDepends lists packages required to build the package, on any architecture.
	BuildDepends dependency.Dependency `debian:"Build-Depends,omitempty" json:"Build-Depends,omitzero"`
	// BuildDependsIndep lists packages required to build the architecture independent binary packages.
	BuildDependsIndep dependency.Dependency `debian:"Build-Depends-Indep,omitempty" json:"Build-Depends-Indep,omitzero"`
	// BuildDependsArch lists packages required to build the architecture dependent binary packages.
	BuildDependsArch dependency.Dependency `debian:"Build-Depends-Arch,omitempty" json:"Build-Depends-Arch,omitzero"`
	// BuildConflicts lists packages that must not be installed while the package is built.
	BuildConflicts dependency.Dependency `debian:"Build-Conflicts,omitempty" json:"Build-Conflicts,omitzero"`
	// BuildConflictsIndep lists packages that must not be installed while the architecture independent binary packages are built.
	BuildConflictsIndep dependency.Dependency `debian:"Build-Conflicts-Indep,omitempty" json:"Build-Conflicts-Indep,omitzero"`
	// BuildConflictsArch lists packages that must not be installed while the architecture dependent binary packages are built.
	BuildConflictsArch dependency.Dependency `debian:"Build-Conflicts-Arch,omitempty" json:"Build-Conflicts-Arch,omitzero"`
	// PackageList lists the binary packages built from this source, one per line,
	// each with its package type, section, priority and build profile fields.
	PackageList list.NewLineDelimited[string] `debian:"Package-List,omitempty" json:"Package-List,omitzero"`
	// ChecksumsSha1 lists the files of the source package with their SHA-1 checksums.
	ChecksumsSha1 list.NewLineDelimited[filehash.FileHash] `debian:"Checksums-Sha1,omitempty" json:"Checksums-Sha1,omitzero"`
	// ChecksumsSha256 lists the files of the source package with their SHA-256 checksums.
	ChecksumsSha256 list.NewLineDelimited[filehash.FileHash] `debian:"Checksums-Sha256,omitempty" json:"Checksums-Sha256,omitzero"`
	// ChecksumsSha512 lists the files of the source package with their SHA-512 checksums.
	ChecksumsSha512 list.NewLineDelimited[filehash.FileHash] `debian:"Checksums-Sha512,omitempty" json:"Checksums-Sha512,omitzero"`
	// Files lists the files of the source package with their MD5 checksums.
	Files list.NewLineDelimited[filehash.FileHash] `debian:"Files" json:"Files"`
}
