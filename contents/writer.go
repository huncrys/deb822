// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package contents

import (
	"fmt"
	"io"
	"strings"
)

// DefaultPadding is the column the path is padded to, matching dak's binary
// Contents writer ("%-55s %s").
const DefaultPadding = 55

// Writer writes a Contents index. Entries are written in the order given;
// sorting them by path, and merging the package lists of a shared path, is the
// caller's job.
type Writer struct {
	w   io.Writer
	pad string
	sep string
	buf []byte
}

// WriterOption configures the layout of the written index.
type WriterOption func(*Writer)

// WithPadding sets the column the path is padded to with spaces. A width of 0
// disables padding, leaving the separator alone between the columns.
func WithPadding(width int) WriterOption {
	return func(w *Writer) {
		if width < 0 {
			width = 0
		}

		w.pad = strings.Repeat(" ", width)
	}
}

// WithTabSeparator separates the columns with a tab rather than a space.
// Combined with WithPadding(0) this is the layout of dak's Contents-source
// indices ("%s\t%s").
func WithTabSeparator() WriterOption {
	return func(w *Writer) {
		w.sep = "\t"
	}
}

// NewWriter returns a Writer writing an uncompressed Contents index to w. With
// no options it writes dak's binary layout: the path padded to DefaultPadding
// columns, then a single space, then the comma separated package list.
func NewWriter(w io.Writer, opts ...WriterOption) *Writer {
	cw := &Writer{
		w:   w,
		pad: strings.Repeat(" ", DefaultPadding),
		sep: " ",
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(cw)
	}

	return cw
}

// Write writes one entry. Entries that could not be read back unambiguously
// are rejected rather than written.
func (w *Writer) Write(e Entry) error {
	if err := validate(e); err != nil {
		return err
	}

	w.buf = append(w.buf[:0], e.Path...)
	if n := len(e.Path); n < len(w.pad) {
		w.buf = append(w.buf, w.pad[n:]...)
	}
	w.buf = append(w.buf, w.sep...)

	for i, name := range e.Packages {
		if i > 0 {
			w.buf = append(w.buf, ',')
		}
		w.buf = append(w.buf, name...)
	}
	w.buf = append(w.buf, '\n')

	if _, err := w.w.Write(w.buf); err != nil {
		return fmt.Errorf("failed to write contents entry: %w", err)
	}

	return nil
}

// WriteAll writes a batch of entries.
func (w *Writer) WriteAll(entries []Entry) error {
	for _, e := range entries {
		if err := w.Write(e); err != nil {
			return err
		}
	}

	return nil
}

// validate rejects entries whose columns would not survive a round trip: the
// path is delimited by whitespace and the package list by commas, so neither
// column may carry those characters at its edges.
func validate(e Entry) error {
	if e.Path == "" || strings.TrimLeft(e.Path, " \t") != e.Path ||
		strings.TrimRight(e.Path, " \t") != e.Path ||
		strings.ContainsAny(e.Path, "\r\n") {
		return fmt.Errorf("%w: %q", ErrInvalidPath, e.Path)
	}

	if len(e.Packages) == 0 {
		return fmt.Errorf("%w: no packages for %q", ErrInvalidPackageList, e.Path)
	}

	for _, name := range e.Packages {
		if name == "" || strings.ContainsAny(name, ", \t\r\n") {
			return fmt.Errorf("%w: %q", ErrInvalidPackageList, name)
		}
	}

	return nil
}
