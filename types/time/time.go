// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package time

import (
	"fmt"
	"strings"

	stdtime "time"
)

// layouts are the date formats accepted when parsing, tried in order.
//
// Debian archives are not consistent: the canonical form is RFC1123 with a
// zone abbreviation, but numeric offsets, single-digit days and weekday-less
// dates all occur in the wild.
var layouts = []string{
	stdtime.RFC1123,                  // "Mon, 02 Jan 2006 15:04:05 MST"
	stdtime.RFC1123Z,                 // "Mon, 02 Jan 2006 15:04:05 -0700"
	"Mon, 2 Jan 2006 15:04:05 MST",   // single-digit day
	"Mon, 2 Jan 2006 15:04:05 -0700", // single-digit day, numeric zone
	"2 Jan 2006 15:04:05 MST",        // no weekday
	"2 Jan 2006 15:04:05 -0700",      // no weekday, numeric zone
}

// Time is an RFC2822 formatted time.
type Time stdtime.Time

func (t Time) MarshalText() ([]byte, error) {
	return []byte(stdtime.Time(t).Format(stdtime.RFC1123)), nil
}

func (t *Time) UnmarshalText(text []byte) error {
	s := string(text)

	var err error

	for _, layout := range layouts {
		var parsed stdtime.Time

		parsed, err = stdtime.Parse(layout, s)
		if err == nil {
			// time.Parse attaches time.Local, name and all, whenever a numeric
			// offset happens to match the host's local zone at that instant.
			// Re-anchoring in an unnamed fixed zone drops the borrowed name, so
			// MarshalText renders the offset numerically on every machine.
			if strings.Contains(layout, "-0700") {
				_, offset := parsed.Zone()
				parsed = parsed.In(stdtime.FixedZone("", offset))
			}

			*t = Time(parsed)

			return nil
		}
	}

	return fmt.Errorf("failed to parse time %q: %w", s, err)
}
