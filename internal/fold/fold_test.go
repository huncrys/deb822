// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package fold_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822/internal/fold"
)

func TestValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "empty",
			value:    "",
			expected: "",
		},
		{
			name:     "single line",
			value:    "a short description",
			expected: "a short description",
		},
		{
			name:     "continuation lines",
			value:    "short\nfirst\nsecond",
			expected: "short\n first\n second",
		},
		{
			name:     "blank line becomes a lone dot",
			value:    "short\nfirst\n\nsecond",
			expected: "short\n first\n .\n second",
		},
		{
			name:     "trailing newline is trimmed",
			value:    "short\nfirst\n",
			expected: "short\n first",
		},
		{
			name:     "indentation under the fold is kept",
			value:    "short\n  indented",
			expected: "short\n   indented",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, fold.Value(test.value))
		})
	}
}
