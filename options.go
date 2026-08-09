// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package deb822

import "errors"

// Sentinel errors returned (wrapped, with the offending line or field name
// attached) by the parser. Use errors.Is to test for them.
var (
	// ErrInvalidFieldName is returned in strict mode when a field name does
	// not conform to Debian Policy 5.1.
	ErrInvalidFieldName = errors.New("invalid field name")

	// ErrDuplicateField is returned in strict mode when a paragraph repeats a
	// field. Field names are compared case-insensitively.
	ErrDuplicateField = errors.New("duplicate field")

	// ErrCommentNotAllowed is returned when a comment line is encountered but
	// the comment policy in effect forbids comments.
	ErrCommentNotAllowed = errors.New("comment not allowed")

	// ErrUnexpectedContinuation is returned when a continuation line appears
	// before any field has been seen in the current paragraph. This check is
	// always on, in both lenient and strict mode.
	ErrUnexpectedContinuation = errors.New("unexpected continuation line")
)

// readerOptions holds the resolved parser configuration.
type readerOptions struct {
	// strict enables Debian Policy 5.1 field-name validation and rejects
	// duplicate fields within a paragraph.
	strict bool

	// comments overrides the comment policy. When nil, comments are allowed
	// unless strict mode is enabled.
	comments *bool
}

// allowComments reports whether comment lines are accepted.
func (o readerOptions) allowComments() bool {
	if o.comments != nil {
		return *o.comments
	}

	return !o.strict
}

// ReaderOption configures the behaviour of a StanzaReader.
type ReaderOption func(*readerOptions)

// WithStrict enables strict parsing: field names are validated against Debian
// Policy 5.1, duplicate fields within a paragraph are rejected, and comment
// lines are rejected unless WithComments(true) is also given.
func WithStrict() ReaderOption {
	return func(o *readerOptions) {
		o.strict = true
	}
}

// WithComments overrides the comment policy independently of strict mode.
// WithComments(true) accepts comment lines even in strict mode (as used by
// debian/control), WithComments(false) rejects them even in lenient mode.
func WithComments(allow bool) ReaderOption {
	return func(o *readerOptions) {
		o.comments = &allow
	}
}

// newReaderOptions resolves a list of options into a readerOptions value.
func newReaderOptions(opts []ReaderOption) readerOptions {
	var resolved readerOptions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&resolved)
	}

	return resolved
}

// validFieldName reports whether name is a valid field name per Debian Policy
// 5.1: it must be non-empty, must not begin with '#' or '-', and must consist
// solely of US-ASCII characters in the range 0x21 to 0x7E excluding ':'.
func validFieldName(name string) bool {
	if name == "" {
		return false
	}

	if name[0] == '#' || name[0] == '-' {
		return false
	}

	for i := 0; i < len(name); i++ {
		if c := name[i]; c < 0x21 || c > 0x7E || c == ':' {
			return false
		}
	}

	return true
}
