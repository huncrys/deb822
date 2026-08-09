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

// ChangesFileHash is an entry found in the Files field of a Debian .changes
// file. Unlike the Release file variant it carries the archive section and
// priority of the referenced file:
//
//	<hash> <size> <section> <priority> <filename>
type ChangesFileHash struct {
	Hash     string
	Size     int64
	Section  string
	Priority string
	Filename string
}

func (h ChangesFileHash) String() string {
	return fmt.Sprintf("%s %d %s %s %s", h.Hash, h.Size, h.Section, h.Priority, h.Filename)
}

func (h ChangesFileHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *ChangesFileHash) UnmarshalText(text []byte) error {
	line := string(text)

	hash, rest, ok := cutField(strings.TrimLeft(line, " "))
	if !ok {
		return fmt.Errorf("missing size field in changes file hash entry %q", line)
	}

	sizeStr, rest, ok := cutField(rest)
	if !ok {
		return fmt.Errorf("missing section field in changes file hash entry %q", line)
	}

	section, rest, ok := cutField(rest)
	if !ok {
		return fmt.Errorf("missing priority field in changes file hash entry %q", line)
	}

	priority, filename, ok := cutField(rest)
	if !ok {
		return fmt.Errorf("missing filename field in changes file hash entry %q", line)
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid size in changes file hash entry %q: %w", line, err)
	}

	if filename == "" {
		return fmt.Errorf("missing filename field in changes file hash entry %q", line)
	}

	h.Hash = hash
	h.Size = size
	h.Section = section
	h.Priority = priority
	h.Filename = filename

	return nil
}
