// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package filehash

import (
	"fmt"
	"strconv"
	"strings"
)

// FileHash is an entry found in a Debian Release file
type FileHash struct {
	Hash     string
	Size     int64
	Filename string
}

func (h FileHash) String() string {
	return fmt.Sprintf("%s %d %s", h.Hash, h.Size, h.Filename)
}

func (h FileHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *FileHash) UnmarshalText(text []byte) error {
	line := string(text)

	hash, rest, ok := cutField(strings.TrimLeft(line, " "))
	if !ok {
		return fmt.Errorf("missing size field in file hash entry %q", line)
	}

	sizeStr, filename, ok := cutField(rest)
	if !ok {
		return fmt.Errorf("missing filename field in file hash entry %q", line)
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid size in file hash entry %q: %w", line, err)
	}

	if filename == "" {
		return fmt.Errorf("missing filename field in file hash entry %q", line)
	}

	h.Hash = hash
	h.Size = size
	h.Filename = filename

	return nil
}

// cutField splits off the first space delimited field of s. The returned
// remainder has any repeated separator spaces stripped from its front, so that
// the next field (or a verbatim trailing filename) starts at its first
// character. Interior spaces of the remainder are left untouched.
func cutField(s string) (field, rest string, ok bool) {
	field, rest, ok = strings.Cut(s, " ")
	if !ok {
		return field, "", false
	}

	return field, strings.TrimLeft(rest, " "), true
}
