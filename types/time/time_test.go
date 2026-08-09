// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package time_test

import (
	"testing"

	stdtime "time"

	"oaklab.hu/debian/deb822/types/time"

	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	t.Run("MarshalText", func(t *testing.T) {
		tm := time.Time(stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC))

		text, err := tm.MarshalText()
		require.NoError(t, err)

		require.Equal(t, "Sat, 10 Feb 2024 11:07:25 UTC", string(text))
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		tests := []struct {
			name       string
			text       string
			wantUTC    stdtime.Time
			wantOffset int
			wantZone   string
		}{
			{
				name:       "RFC1123",
				text:       "Sat, 10 Feb 2024 11:07:25 UTC",
				wantUTC:    stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC),
				wantOffset: 0,
				wantZone:   "UTC",
			},
			{
				name:       "RFC1123Z",
				text:       "Sat, 10 Feb 2024 11:07:25 +0000",
				wantUTC:    stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC),
				wantOffset: 0,
				// "+0000" is accepted by the zone *name* layout, which keeps the
				// literal as the name. Rendering it back is still "+0000".
				wantZone: "+0000",
			},
			{
				name:       "single digit day with zone name",
				text:       "Sat, 2 Jul 2016 05:20:50 UTC",
				wantUTC:    stdtime.Date(2016, stdtime.July, 2, 5, 20, 50, 0, stdtime.UTC),
				wantOffset: 0,
				wantZone:   "UTC",
			},
			{
				name:       "single digit day with numeric zone",
				text:       "Sat, 2 Jul 2016 05:20:50 +0100",
				wantUTC:    stdtime.Date(2016, stdtime.July, 2, 4, 20, 50, 0, stdtime.UTC),
				wantOffset: 3600,
				wantZone:   "",
			},
			{
				name:       "no weekday with zone name",
				text:       "2 Jul 2016 05:20:50 UTC",
				wantUTC:    stdtime.Date(2016, stdtime.July, 2, 5, 20, 50, 0, stdtime.UTC),
				wantOffset: 0,
				wantZone:   "UTC",
			},
			{
				name:       "no weekday with numeric zone",
				text:       "2 Jul 2016 05:20:50 +0100",
				wantUTC:    stdtime.Date(2016, stdtime.July, 2, 4, 20, 50, 0, stdtime.UTC),
				wantOffset: 3600,
				wantZone:   "",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var tm time.Time
				require.NoError(t, tm.UnmarshalText([]byte(tc.text)))

				parsed := stdtime.Time(tm)
				require.Equal(t, tc.wantUTC, parsed.UTC())

				zone, offset := parsed.Zone()
				require.Equal(t, tc.wantOffset, offset)
				require.Equal(t, tc.wantZone, zone)
			})
		}
	})

	t.Run("numeric zone does not borrow the local zone name", func(t *testing.T) {
		// Pin the host zone to the one the input carries: that is exactly the
		// case where time.Parse hands back time.Local instead of a fixed zone.
		// time.Local is process global, so this cannot run in parallel.
		saved := stdtime.Local
		t.Cleanup(func() { stdtime.Local = saved })

		stdtime.Local = stdtime.FixedZone("CET", 3600)

		var tm time.Time
		require.NoError(t, tm.UnmarshalText([]byte("Thu, 07 Mar 2024 12:34:56 +0100")))

		zone, offset := stdtime.Time(tm).Zone()
		require.Empty(t, zone)
		require.Equal(t, 3600, offset)

		text, err := tm.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "Thu, 07 Mar 2024 12:34:56 +0100", string(text))

		// A named zone keeps its name: apt only accepts GMT, UTC, Z or a zero
		// numeric offset in Release files.
		var named time.Time
		require.NoError(t, named.UnmarshalText([]byte("Sat, 10 Feb 2024 11:07:25 UTC")))

		namedZone, namedOffset := stdtime.Time(named).Zone()
		require.Equal(t, "UTC", namedZone)
		require.Equal(t, 0, namedOffset)

		namedText, err := named.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "Sat, 10 Feb 2024 11:07:25 UTC", string(namedText))
	})

	t.Run("UnmarshalText InRelease Date", func(t *testing.T) {
		// The literal Date field value from testdata/InRelease.
		text := "Sat, 10 Feb 2024 11:07:25 UTC"

		var tm time.Time
		require.NoError(t, tm.UnmarshalText([]byte(text)))

		require.Equal(t, stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC), stdtime.Time(tm).UTC())
	})

	t.Run("RoundTrip", func(t *testing.T) {
		text := []byte("Sat, 10 Feb 2024 11:07:25 UTC")

		var tm time.Time
		require.NoError(t, tm.UnmarshalText(text))

		out, err := tm.MarshalText()
		require.NoError(t, err)

		require.Equal(t, text, out)
	})

	t.Run("UnmarshalText invalid", func(t *testing.T) {
		for _, text := range []string{
			"invalid date string",
			"",
			"Sat, 10 Foo 2024 11:07:25 UTC",
			"2024-02-10T11:07:25Z",
		} {
			var tm time.Time

			err := tm.UnmarshalText([]byte(text))
			require.Error(t, err)
			require.Contains(t, err.Error(), text)
		}
	})
}
