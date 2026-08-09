# deb822

Go SerDes library for Debian deb822 control files. Forked from
`github.com/dpeckett/deb822` (itself derived from `paultag/go-debian`) and
heavily diverged since.

## dpkg and apt are the reference, not the spec

Where Debian Policy is ambiguous or the tools deviate from it, this library
matches what dpkg and apt actually do. Several behaviours look like bugs but
are deliberate, pinned by tests, and must not be "fixed":

- Upstream versions not starting with a digit are accepted - dpkg only warns.
  Empty revisions (`1.0-`), empty upstream versions (`1:-1`) and non-numeric
  epochs are rejected; epochs are capped at INT_MAX (dpkg parses them as int).
- The legacy dependency operators `<` and `>` parse as `<=` and `>=`.
- `arch.Arch` models dpkg's four-component tuple (abi-libc-os-cpu). Wildcards
  left-pad with `any`; `String()` emits dpkg-canonical short forms and only
  shortens when the result reparses to the identical tuple.
- Dates decoded from a numeric zone offset are re-anchored in
  `time.FixedZone("", offset)`. This is not redundant: `time.Parse` attaches
  `time.Local` (name and all) when the offset matches the host zone, and a
  zone *name* in the output is invalid downstream - dpkg requires a numeric
  `[-+]\d{4}` in .changes dates, and apt accepts only `GMT`/`UTC`/`Z` or a
  zero numeric offset in Release dates. Named `UTC` input must keep
  round-tripping as `UTC`; that is what real Release files carry.

## Encoding is byte-stable

Struct field declaration order in `types/` is the wire order - the stanza
walker iterates fields in declaration order, and `TestEncodeIsStable` pins
that decode/encode reaches a fixed point. Reordering struct fields changes
output for every consumer; treat it as a breaking change. The `testdata/`
files are real Debian archive snapshots and serve as ground truth.

## Struct tags

Serialization uses the `debian:` tag (`debian:"Field-Name[,omitempty][,inline]"`),
falling back to the `json:` tag's name part, then the Go field name.
`omitempty` is operative, not decorative: text-empty values are always
skipped, and `omitempty` additionally skips Go-zero values (a zero `Size`
would otherwise encode, and the list types marshal a leading newline even
when empty). The `json:` tags exist solely for `encoding/json` and carry
`omitzero` on optional fields deliberately - `TestJSONHasNoEmptyValues`
guards that a marshalled `Package` never emits empty strings, empty lists or
nulls.

## Lenient by default

Parsing is lenient unless `WithStrict()`/`WithComments()` are passed, and the
defaults are frozen: `TestLenientDefaultsUnchanged` decodes the full testdata
with no options and must keep passing. Policy validation goes behind options,
never into the default path.

## Copyright headers

Files derived from go-debian keep the Damian Peckett / Paul Tagliamonte
attribution blocks. Newly authored files carry only
`Copyright (C) <year> Kristof Bach <crys@crys.hu>` with the MPL-2.0 notice.
Match whichever applies when adding a file; never blanket-rewrite existing
headers.
