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
	"oaklab.hu/debian/deb822/types/dependency"
	"oaklab.hu/debian/deb822/types/filehash"
	"oaklab.hu/debian/deb822/types/list"
	"oaklab.hu/debian/deb822/types/version"
)

// Source is one stanza of a repository's Sources index: the metadata of a
// single source package as an archive publishes it, which is the .dsc of that
// source package plus the archive's own bookkeeping (Directory, Priority,
// Section, Extra-Source-Only).
//
// Beware the name collision. This types.Source is a Sources index entry. It is
// unrelated to dependency.Source, which models the "Source:" pseudo-field of a
// *binary* package stanza - a name, optionally with a version in parentheses,
// pointing at the source package a binary was built from. A Package carries the
// latter in its Source field; a Sources index carries the former.
type Source struct {
	// Package is the name of the source package.
	Package string `debian:"Package" json:"Package"`
	// Format is the source package format, such as "3.0 (quilt)".
	Format string `debian:"Format" json:"Format"`
	// Binary lists the binary packages this source package builds.
	Binary list.CommaDelimited[string] `debian:"Binary,omitempty" json:"Binary,omitzero"`
	// Architecture lists the architectures the source package can be built for.
	Architecture list.SpaceDelimited[arch.Arch] `debian:"Architecture,omitempty" json:"Architecture,omitzero"`
	// Version is the version of the source package.
	Version version.Version `debian:"Version" json:"Version"`
	// Priority is the default priority of the binary packages built from this source.
	Priority string `debian:"Priority,omitempty" json:"Priority,omitzero"`
	// Section is the default archive section of the binary packages built from this source.
	Section string `debian:"Section,omitempty" json:"Section,omitzero"`
	// Maintainer is the name and email address of the person or organization responsible for the package.
	Maintainer string `debian:"Maintainer" json:"Maintainer"`
	// Uploaders lists co-maintainers allowed to upload the package.
	Uploaders list.CommaDelimited[string] `debian:"Uploaders,omitempty" json:"Uploaders,omitzero"`
	// OriginalMaintainer records the maintainer of the package before a derivative distribution took it over.
	OriginalMaintainer string `debian:"Original-Maintainer,omitempty" json:"Original-Maintainer,omitzero"`
	// StandardsVersion is the version of the Debian Policy the package claims to comply with.
	StandardsVersion string `debian:"Standards-Version,omitempty" json:"Standards-Version,omitzero"`
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
	// Testsuite names the automatic test suites the package ships, such as "autopkgtest".
	Testsuite string `debian:"Testsuite,omitempty" json:"Testsuite,omitzero"`
	// TestsuiteTriggers lists the binary packages whose upload should trigger a run of the test suite.
	TestsuiteTriggers string `debian:"Testsuite-Triggers,omitempty" json:"Testsuite-Triggers,omitzero"`
	// Homepage is the URL of the upstream project's homepage.
	Homepage string `debian:"Homepage,omitempty" json:"Homepage,omitzero"`
	// Description is the description of the source package.
	Description string `debian:"Description,omitempty" json:"Description,omitzero"`
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
	// ExtraSourceOnly marks a source package that is only in the archive because
	// another source package's build depends on it; it is not a candidate for
	// building on its own.
	ExtraSourceOnly *boolean.Boolean `debian:"Extra-Source-Only,omitempty" json:"Extra-Source-Only,omitzero"`
	// Directory is the pool directory holding the files of the source package,
	// relative to the root of the repository.
	Directory string `debian:"Directory" json:"Directory"`
	// PackageList lists the binary packages built from this source, one per line,
	// each with its package type, section, priority and build profile fields.
	PackageList list.NewLineDelimited[string] `debian:"Package-List,omitempty" json:"Package-List,omitzero"`
	// Files lists the files of the source package with their MD5 checksums.
	Files list.NewLineDelimited[filehash.FileHash] `debian:"Files,omitempty" json:"Files,omitzero"`
	// ChecksumsSha1 lists the files of the source package with their SHA-1 checksums.
	ChecksumsSha1 list.NewLineDelimited[filehash.FileHash] `debian:"Checksums-Sha1,omitempty" json:"Checksums-Sha1,omitzero"`
	// ChecksumsSha256 lists the files of the source package with their SHA-256 checksums.
	ChecksumsSha256 list.NewLineDelimited[filehash.FileHash] `debian:"Checksums-Sha256,omitempty" json:"Checksums-Sha256,omitzero"`
	// ChecksumsSha512 lists the files of the source package with their SHA-512 checksums.
	ChecksumsSha512 list.NewLineDelimited[filehash.FileHash] `debian:"Checksums-Sha512,omitempty" json:"Checksums-Sha512,omitzero"`
}
