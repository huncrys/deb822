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
 * Copyright (c) Paul R. Tagliamonte <paultag@debian.org>, 2015
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package dependency

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"oaklab.hu/debian/deb822"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/version"
)

// Parse a string into a Dependency object. The input should look something
// like "foo, bar | baz".
func Parse(in string) (Dependency, error) {
	var result Dependency
	return result, parseDependency(in, &result)
}

// MustParse is a helper function to wrap Parse and panic on error.
func MustParse(in string) Dependency {
	result, err := Parse(in)
	if err != nil {
		panic(err)
	}
	return result
}

func parseDependency(in string, ret *Dependency) error {
	reader := deb822.NewRuneReader(bytes.NewReader([]byte(in)))

	reader.DiscardSpace()

	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch peek {
		case ',': /* Next relation set */
			reader.DiscardRune()
			reader.DiscardSpace()

			continue
		}

		if err := parseRelation(reader, ret); err != nil {
			return err
		}
	}
}

func parseRelation(reader *deb822.RuneReader, dependency *Dependency) error {
	reader.DiscardSpace()

	ret := &Relation{Possibilities: []Possibility{}}

	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				goto PARSE_RELATION_DONE
			}
			return err
		}

		switch peek {
		case ',': /* Done with this relation! yay */
			goto PARSE_RELATION_DONE
		case '|': /* Next Possibility */
			reader.DiscardRune()
			reader.DiscardSpace()
			continue
		}
		if err := parsePossibility(reader, ret); err != nil {
			return err
		}
	}

PARSE_RELATION_DONE:
	dependency.Relations = append(dependency.Relations, *ret)
	return nil
}

func parsePossibility(reader *deb822.RuneReader, relation *Relation) error {
	reader.DiscardSpace()

	peek, _, err := reader.PeekRune()
	if err != nil {
		return err
	}

	if peek == '$' {
		/* OK, nice. So, we've got a substvar. Let's eat it. */
		return parseSubstvar(reader, relation)
	}

	/* Otherwise, let's punt and build it up ourselves. */

	ret := &Possibility{
		Name:          "",
		Version:       nil,
		Architectures: &ArchSet{Architectures: []arch.Arch{}},
		StageSets:     []StageSet{},
		Substvar:      false,
	}

	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				goto PARSE_POSSIBILITY_DONE
			}
			return err
		}

		switch peek {
		case ':':
			reader.DiscardRune()
			err := parseMultiarch(reader, ret)
			if err != nil {
				return err
			}
			continue
		case ' ', '(':
			err := parsePossibilityControllers(reader, ret)
			if err != nil {
				return err
			}
			continue
		case ',', '|': /* I'm out! */
			goto PARSE_POSSIBILITY_DONE
		}
		/* Not a control, let's append */
		reader.DiscardRune()
		ret.Name += string(peek)
	}

PARSE_POSSIBILITY_DONE:
	if ret.Name == "" {
		return nil // e.g. trailing comma in Build-Depends
	}
	relation.Possibilities = append(relation.Possibilities, *ret)
	return nil
}

func parseSubstvar(reader *deb822.RuneReader, relation *Relation) error {
	reader.DiscardSpace()
	reader.DiscardRune() /* Assert ch == '$' */
	reader.DiscardRune() /* Assert ch == '{' */

	ret := &Possibility{
		Name:     "",
		Version:  nil,
		Substvar: true,
	}

	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Substvar finished: %w", err)
			}
			return err
		}

		if peek == '}' {
			reader.DiscardRune()
			relation.Possibilities = append(relation.Possibilities, *ret)
			return nil
		}
		next, _, _ := reader.ReadRune()
		ret.Name += string(next)
	}
}

func parseMultiarch(reader *deb822.RuneReader, possi *Possibility) error {
	name := ""
	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				goto PARSE_MULTIARCH_DONE
			}
			return err
		}

		switch peek {
		case ',', '|', ' ', '(', '[', '<':
			goto PARSE_MULTIARCH_DONE
		default:
			reader.DiscardRune()
			name += string(peek)
		}
	}

PARSE_MULTIARCH_DONE:
	archObj, err := arch.Parse(name)
	if err != nil {
		return err
	}
	possi.Arch = &archObj
	return nil
}

func parsePossibilityControllers(reader *deb822.RuneReader, possi *Possibility) error {
	for {
		reader.DiscardSpace()
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch peek {
		case ',', '|':
			return nil
		case '(':
			if possi.Version != nil {
				return errors.New("only one Version relation per Possibility")
			}
			err := parsePossibilityVersion(reader, possi)
			if err != nil {
				return err
			}
			continue
		case '[':
			if len(possi.Architectures.Architectures) != 0 {
				return errors.New("only one Arch relation per Possibility")
			}
			err := parsePossibilityArchs(reader, possi)
			if err != nil {
				return err
			}
			continue
		case '<':
			err := parsePossibilityStageSet(reader, possi)
			if err != nil {
				return err
			}
			continue
		}
		return fmt.Errorf("trailing garbage in Possibility: %s", string(peek))
	}
}

func parsePossibilityVersion(reader *deb822.RuneReader, possi *Possibility) error {
	reader.DiscardSpace()
	reader.DiscardRune() /* mandated to be ( */
	version := VersionRelation{}

	err := parsePossibilityOperator(reader, &version)
	if err != nil {
		return err
	}

	err = parsePossibilityNumber(reader, &version)
	if err != nil {
		return err
	}

	_, _, _ = reader.ReadRune() /* mandated to be ) */
	possi.Version = &version
	return nil
}

func parsePossibilityOperator(reader *deb822.RuneReader, version *VersionRelation) error {
	reader.DiscardSpace()
	leader, _, err := reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("reached EOF before Operator finished: %w", err)
		}
		return err
	}

	if leader == '=' {
		version.Operator = "="
		return nil
	}

	secondary, _, err := reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("reached EOF before Operator finished: %w", err)
		}
		return err
	}

	operator := string([]rune{leader, secondary})

	switch operator {
	case ">=", "<=", "<<", ">>":
		version.Operator = operator
		return nil
	}

	return fmt.Errorf("unknown Operator in Possibility Version modifier: %s", operator)
}

func parsePossibilityNumber(reader *deb822.RuneReader, version *VersionRelation) error {
	reader.DiscardSpace()
	versionStr := ""
	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Number finished: %w", err)
			}
			return err
		}

		if peek == ')' {
			return version.Version.UnmarshalText([]byte(versionStr))
		}

		reader.DiscardRune()
		versionStr += string(peek)
	}
}

func parsePossibilityArchs(reader *deb822.RuneReader, possi *Possibility) error {
	reader.DiscardSpace()
	reader.DiscardRune() /* Assert ch == '[' */

	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Arch list finished: %w", err)
			}
			return err
		}

		if peek == ']' {
			reader.DiscardRune()
			return nil
		}

		if err := parsePossibilityArch(reader, possi); err != nil {
			return err
		}
	}
}

func parsePossibilityArch(reader *deb822.RuneReader, possi *Possibility) error {
	reader.DiscardSpace()
	name := ""

	peek, _, err := reader.PeekRune()
	if err != nil {
		return err
	}
	hasNot := peek == '!'
	if hasNot {
		reader.DiscardRune() // '!'
	}
	if len(possi.Architectures.Architectures) == 0 {
		possi.Architectures.Not = hasNot
	} else if possi.Architectures.Not != hasNot {
		return errors.New("cannot mix negated and non-negated architectures")
	}

	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Arch list finished: %w", err)
			}
			return err
		}

		switch peek {
		case '!':
			return errors.New("you can only negate whole blocks")
		case ']', ' ': /* Let our parent deal with both of these */
			archObj, err := arch.Parse(name)
			if err != nil {
				return err
			}
			possi.Architectures.Architectures = append(possi.Architectures.Architectures, archObj)
			return nil
		}
		reader.DiscardRune()
		name += string(peek)
	}
}

func parsePossibilityStageSet(reader *deb822.RuneReader, possi *Possibility) error {
	reader.DiscardSpace()
	reader.DiscardRune() /* Assert ch == '<' */

	stageSet := StageSet{}
	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before StageSet finished: %w", err)
			}
			return err
		}

		if peek == '>' {
			reader.DiscardRune()
			possi.StageSets = append(possi.StageSets, stageSet)
			return nil
		}

		if err := parsePossibilityStage(reader, &stageSet); err != nil {
			return err
		}
	}
}

func parsePossibilityStage(reader *deb822.RuneReader, stageSet *StageSet) error {
	reader.DiscardSpace()

	stage := Stage{}
	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Stage finished: %w", err)
			}
			return err
		}

		switch peek {
		case '!':
			reader.DiscardRune()
			if stage.Not {
				return errors.New("double negation in Stage is not permitted")
			}
			stage.Not = !stage.Not
			continue
		case '>', ' ': /* Let our parent deal with both of these */
			stageSet.Stages = append(stageSet.Stages, stage)
			return nil
		}
		reader.DiscardRune()
		stage.Name += string(peek)
	}
}

func parseSource(in string, ret *Source) error {
	reader := deb822.NewRuneReader(bytes.NewReader([]byte(in)))
	reader.DiscardSpace()

	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if peek == ' ' || peek == '(' {
			goto PARSE_NAME_DONE
		}

		reader.DiscardRune()
		ret.Name += string(peek)
	}

PARSE_NAME_DONE:
	reader.DiscardSpace()

	next, _, err := reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	if next != '(' {
		return fmt.Errorf("expected '(', got '%c'", next)
	}

	versionStr := ""
	for {
		peek, _, err := reader.PeekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Number finished: %w", err)
			}
			return err
		}

		if peek == ')' {
			reader.DiscardRune()

			parsed, err := version.Parse(versionStr)
			if err != nil {
				return err
			}
			ret.Version = &parsed
			break
		}

		reader.DiscardRune()
		versionStr += string(peek)
	}

	if _, err := reader.ReadByte(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing garbage after Source version: %w", err)
	}

	return nil
}
