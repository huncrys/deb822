// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package boolean_test

import (
	"testing"

	"github.com/dpeckett/deb822/types/boolean"
	"github.com/stretchr/testify/require"
)

func TestBoolean(t *testing.T) {
	t.Run("MarshalText", func(t *testing.T) {
		t.Run("true", func(t *testing.T) {
			b := boolean.Boolean(true)

			text, err := b.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "yes", string(text))
		})

		t.Run("false", func(t *testing.T) {
			b := boolean.Boolean(false)

			text, err := b.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "no", string(text))
		})
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		t.Run("yes", func(t *testing.T) {
			var b boolean.Boolean
			require.NoError(t, b.UnmarshalText([]byte("yes")))

			require.True(t, bool(b))
		})

		t.Run("no", func(t *testing.T) {
			var b boolean.Boolean
			require.NoError(t, b.UnmarshalText([]byte("no")))

			require.False(t, bool(b))
		})

		t.Run("true", func(t *testing.T) {
			var b boolean.Boolean
			require.NoError(t, b.UnmarshalText([]byte("true")))

			require.True(t, bool(b))
		})

		t.Run("false", func(t *testing.T) {
			var b boolean.Boolean
			require.NoError(t, b.UnmarshalText([]byte("false")))

			require.False(t, bool(b))
		})
	})
}
