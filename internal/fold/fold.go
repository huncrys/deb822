// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

// Package fold turns a decoded field value back into the multi-line form a
// deb822 document carries on the wire. It exists so that the stanza writer and
// anything that has to hash a field as published share one implementation.
package fold

import "strings"

// Value folds a field value into its on-wire form: every line after the first
// is indented by a single space, a line that would otherwise be blank becomes a
// lone ".", and the padding a fold leaves at the end is trimmed off.
//
// It is the inverse of the unfolding the stanza reader performs on decode.
func Value(value string) string {
	value = strings.ReplaceAll(value, "\n", "\n ")
	value = strings.ReplaceAll(value, "\n \n", "\n .\n")

	return strings.TrimRight(value, "\n ")
}
