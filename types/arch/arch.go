// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Portions of this file are based on code originally from: github.com/paultag/go-debian
 *
 * Copyright (c) Paul R. Tagliamonte <paultag@debian.org>, 2015
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package arch

import (
	"errors"
	"slices"
	"strings"
)

// Default components used to complete a shortened concrete architecture
// string, and to shorten a tuple back into its canonical form.
const (
	defaultABI  = "base"
	defaultLibc = "gnu"
	defaultOS   = "linux"
)

// Arch is a Debian architecture tuple, as defined by dpkg since 1.18.11.
//
// A tuple has four components, written abi-libc-os-cpu, for example
// base-gnu-linux-amd64. Most tuples are written in a shortened form that
// omits the leading components that hold their default value, so the tuple
// above is normally written simply as amd64.
type Arch struct {
	ABI  string // e.g. "base", "eabihf"
	Libc string // e.g. "gnu", "musl"
	OS   string // e.g. "linux", "kfreebsd"
	CPU  string // e.g. "amd64", "all", "any"
}

// IsWildcard reports whether the architecture contains a wildcard component.
func (arch *Arch) IsWildcard() bool {
	if arch.CPU == "all" {
		return false
	}

	return arch.hasAny()
}

// hasAny reports whether any component of the tuple is the "any" wildcard.
func (arch *Arch) hasAny() bool {
	return arch.ABI == "any" || arch.Libc == "any" || arch.OS == "any" || arch.CPU == "any"
}

func (arch *Arch) Is(other *Arch) bool {
	if arch.IsWildcard() && other.IsWildcard() {
		/* We can't compare wildcards to other wildcards. That's just
		* insanity. We always need a concrete arch. Not even going to try. */
		return false
	} else if arch.IsWildcard() {
		/* OK, so we're a wildcard. Let's defer to the other
		* struct to deal with this */
		return other.Is(arch)
	}

	/* "all" is never satisfied by a wildcard, only by a literal "all". */
	return (arch.CPU == other.CPU || (arch.CPU != "all" && other.CPU == "any")) &&
		(arch.OS == other.OS || other.OS == "any") &&
		(arch.Libc == other.Libc || other.Libc == "any") &&
		(arch.ABI == other.ABI || other.ABI == "any")
}

// String returns the canonical, shortest representation of the architecture
// tuple. The result always parses back into the very same tuple.
func (arch Arch) String() string {
	/* The zero value renders as the empty string, not as "---". */
	if arch.ABI == "" && arch.Libc == "" && arch.OS == "" && arch.CPU == "" {
		return ""
	}

	components := []string{arch.ABI, arch.Libc, arch.OS, arch.CPU}

	switch {
	case arch.ABI == "all" && arch.Libc == "all" && arch.OS == "all" && arch.CPU == "all":
		return "all"
	case arch.ABI == "any" && arch.Libc == "any" && arch.OS == "any" && arch.CPU == "any":
		return "any"
	}

	var short string
	if arch.hasAny() {
		short = shortenWildcard(components)
	} else {
		short = shortenConcrete(components)
	}

	/* Only use the shortened form if it parses back into the same tuple,
	* otherwise fall back to spelling the whole tuple out. */
	if round, err := Parse(short); err == nil && round == arch {
		return short
	}

	return strings.Join(components, "-")
}

// shortenWildcard drops leading "any" components for as long as the remainder
// still holds an "any", so that the left-padding rule restores them on parse.
func shortenWildcard(components []string) string {
	for len(components) > 1 && components[0] == "any" && slices.Contains(components[1:], "any") {
		components = components[1:]
	}

	return strings.Join(components, "-")
}

// shortenConcrete drops leading components for as long as they hold their
// default value. Once a component is kept, everything after it is kept too.
func shortenConcrete(components []string) string {
	defaults := []string{defaultABI, defaultLibc, defaultOS}

	i := 0
	for i < len(defaults) && components[i] == defaults[i] {
		i++
	}

	return strings.Join(components[i:], "-")
}

func (arch Arch) MarshalText() ([]byte, error) {
	return []byte(arch.String()), nil
}

func (arch *Arch) UnmarshalText(text []byte) error {
	return parseArchInto(arch, string(text))
}

// Parse an architecture string into an Arch struct.
func Parse(arch string) (Arch, error) {
	var result Arch
	if err := parseArchInto(&result, arch); err != nil {
		return Arch{}, err
	}

	return result, nil
}

// MustParse is like Parse, but panics on error.
func MustParse(arch string) Arch {
	result, err := Parse(arch)
	if err != nil {
		panic(err)
	}
	return result
}

// parseArchInto expands an architecture string into a full four component
// tuple. It always assigns every component, and leaves ret untouched on error.
func parseArchInto(ret *Arch, arch string) error {
	if arch == "" {
		return errors.New("invalid arch string")
	}

	/* May be in the following form:
	* `any` (a wildcard, implicitly any-any-any-any)
	* linux-any (a wildcard, implicitly any-any-linux-any)
	* amd64 (implicitly base-gnu-linux-amd64)
	* kfreebsd-amd64 (implicitly base-gnu-kfreebsd-amd64)
	* musl-linux-amd64 (implicitly base-musl-linux-amd64)
	* eabihf-gnu-linux-armhf */
	components := strings.Split(arch, "-")
	if len(components) > 4 {
		return errors.New("invalid arch string")
	}

	if slices.Contains(components, "any") {
		/* Wildcards are left-padded with "any", following dpkg's
		* Dpkg::Arch::debwildcard_to_debtuple. */
		for len(components) < 4 {
			components = append([]string{"any"}, components...)
		}
	} else {
		switch len(components) {
		case 1:
			if components[0] == "all" {
				components = []string{"all", "all", "all", "all"}
			} else {
				components = []string{defaultABI, defaultLibc, defaultOS, components[0]}
			}
		case 2:
			components = []string{defaultABI, defaultLibc, components[0], components[1]}
		case 3:
			components = append([]string{defaultABI}, components...)
		}
	}

	ret.ABI, ret.Libc, ret.OS, ret.CPU = components[0], components[1], components[2], components[3]

	return nil
}
