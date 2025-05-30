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
 * Copyright (c) 2012 Michael Stapelberg and contributors
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 *     * Neither the name of Michael Stapelberg nor the
 *       names of contributors may be used to endorse or promote products
 *       derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY Michael Stapelberg ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL Michael Stapelberg BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Version represents a dpkg version string.
type Version struct {
	Epoch    uint
	Version  string
	Revision string
}

func (v *Version) Empty() bool {
	return v.Epoch == 0 && v.Version == "" && v.Revision == ""
}

func (v *Version) IsNative() bool {
	return len(v.Revision) == 0
}

func (version Version) MarshalText() ([]byte, error) {
	return []byte(version.String()), nil
}

func (version *Version) UnmarshalText(text []byte) error {
	var err error
	*version, err = Parse(string(text))
	if err != nil {
		return err
	}
	return nil
}

func (v Version) StringWithoutEpoch() string {
	result := v.Version
	if len(v.Revision) > 0 {
		result += "-" + v.Revision
	}
	return result
}

func (v Version) String() string {
	if v.Epoch > 0 {
		return fmt.Sprintf("%d:%s", v.Epoch, v.StringWithoutEpoch())
	}
	return v.StringWithoutEpoch()
}

// Compare compares the two provided Debian versions. It returns 0 if a and b
// are equal, a value < 0 if a is smaller than b and a value > 0 if a is
// greater than b.
func (a Version) Compare(b Version) int {
	if a.Epoch > b.Epoch {
		return 1
	}
	if a.Epoch < b.Epoch {
		return -1
	}

	rc := verrevcmp(a.Version, b.Version)
	if rc != 0 {
		return rc
	}

	return verrevcmp(a.Revision, b.Revision)
}

// Parse returns a Version struct filled with the epoch, version and revision
// specified in input. It verifies the version string as a whole, just like
// dpkg(1), and even returns roughly the same error messages.
func Parse(input string) (Version, error) {
	result := Version{}
	return result, parseInto(&result, input)
}

// MustParse is like Parse, but panics on error.
func MustParse(input string) Version {
	result, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return result
}

func parseInto(result *Version, input string) error {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return errors.New("version string is empty")
	}

	if strings.IndexFunc(trimmed, unicode.IsSpace) != -1 {
		return errors.New("version string has embedded spaces")
	}

	colon := strings.Index(trimmed, ":")
	if colon != -1 {
		epoch, err := strconv.ParseInt(trimmed[:colon], 10, 64)
		if err != nil {
			return fmt.Errorf("epoch: %v", err)
		}
		if epoch < 0 {
			return errors.New("epoch in version is negative")
		}
		result.Epoch = uint(epoch)
	}

	result.Version = trimmed[colon+1:]
	if len(result.Version) == 0 {
		return errors.New("nothing after colon in version number")
	}
	if hyphen := strings.LastIndex(result.Version, "-"); hyphen != -1 {
		result.Revision = result.Version[hyphen+1:]
		result.Version = result.Version[:hyphen]
	}

	if len(result.Version) > 0 && !unicode.IsDigit(rune(result.Version[0])) {
		return errors.New("version number does not start with digit")
	}

	if strings.IndexFunc(result.Version, func(c rune) bool {
		return !cisdigit(c) && !cisalpha(c) && c != '.' && c != '-' && c != '+' && c != '~' && c != ':'
	}) != -1 {
		return errors.New("invalid character in version number")
	}

	if strings.IndexFunc(result.Revision, func(c rune) bool {
		return !cisdigit(c) && !cisalpha(c) && c != '.' && c != '+' && c != '~'
	}) != -1 {
		return errors.New("invalid character in revision number")
	}

	return nil
}

func verrevcmp(a string, b string) int {
	i := 0
	j := 0
	for i < len(a) || j < len(b) {
		var first_diff int
		for (i < len(a) && !cisdigit(rune(a[i]))) ||
			(j < len(b) && !cisdigit(rune(b[j]))) {
			ac := 0
			if i < len(a) {
				ac = order(rune(a[i]))
			}
			bc := 0
			if j < len(b) {
				bc = order(rune(b[j]))
			}
			if ac != bc {
				return ac - bc
			}
			i++
			j++
		}

		for i < len(a) && a[i] == '0' {
			i++
		}
		for j < len(b) && b[j] == '0' {
			j++
		}

		for i < len(a) && cisdigit(rune(a[i])) && j < len(b) && cisdigit(rune(b[j])) {
			if first_diff == 0 {
				first_diff = int(rune(a[i]) - rune(b[j]))
			}
			i++
			j++
		}

		if i < len(a) && cisdigit(rune(a[i])) {
			return 1
		}
		if j < len(b) && cisdigit(rune(b[j])) {
			return -1
		}
		if first_diff != 0 {
			return first_diff
		}
	}
	return 0
}

func cisdigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func cisalpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func order(r rune) int {
	if cisdigit(r) {
		return 0
	}
	if cisalpha(r) {
		return int(r)
	}
	if r == '~' {
		return -1
	}
	if int(r) != 0 {
		return int(r) + 256
	}
	return 0
}
