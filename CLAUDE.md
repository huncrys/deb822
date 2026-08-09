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

## Contents indices are not deb822

`contents/` is the one package that does not parse stanzas. Its invariants come
from the generators, not from a spec:

- Split a line on the **last** whitespace run, never the first. dak pads the
  path to 55 columns with spaces (`%-55s %s`), dak's Contents-source writer uses
  a lone tab, apt-ftparchive pads with tabs, and `testdata/Contents-all` really
  does contain paths with spaces that are exactly 55 columns wide - the
  separator degenerates to a single space there.
- Qualified names are `[[area/]section/]package`; Debian and Ubuntu both emit
  the three-component form, so nothing may assume two.
- `TestWriteIsByteStable` pins that decode/encode of the real dak file
  reproduces it byte for byte. The writer never emits the legacy prose header.
- Compression stays out of the package: both ends take plain streams.

## Changelogs are not deb822 either

`changelog/` is the other non-stanza package. Its invariants also come from the
generators:

- `Entry.Changes` is the body **verbatim** - blank lines, indentation, trailing
  whitespace and all - and the writer replays it untouched. dpkg and `dch` put a
  blank line after the header; goreleaser/chglog (what nFPM uses) does not, and
  indents continuations by three spaces. Neither is canonical, so normalizing
  either would corrupt the other. `TestWriteIsByteStable` pins the chglog file,
  `TestWritePreservesBody` pins the ragged edges.
- Framing is positional: the header is the only line in column 0, the trailer
  the only one starting with `" -- "` (exactly one space). Everything else in an
  entry is body.
- The distribution list may be **empty** (`hello (1.3-6); priority=LOW`) and the
  urgency may be absent or uppercase. dpkg's own regex requires a distribution;
  the archive disagrees, and the archive wins.
- Trailer dates go through `types/time`, so they are re-emitted RFC1123 with a
  numeric zone. hello's space-padded 1990s days therefore do *not* round trip
  byte for byte - `TestReformatIsIdempotent` pins the fixed point instead. Do
  not "fix" this by storing the raw date string. A trailing blank line at EOF is
  dropped the same way (the libc6-*-cross changelogs ship one).
- A date no layout accepts gets one salvage pass - drop the weekday, cut the
  month to three letters - because `dpkg-parsechangelog` never parses this
  field, it hands the string through verbatim. bash is still in the archive with
  `Thur, 19 June 1997 19:13:34 +0100`. Without the salvage one 1997 line fails
  the whole file.
- `ErrNotDebianFormat` (first non-blank line is not a header) is load-bearing
  for aptify: packages without a `changelog.Debian` ship upstream's own file,
  which must be publishable unparsed rather than rejected. Free-form history
  past the last entry goes to `Reader.Trailing()`, never to an error.
- Compression stays out of the package here too.

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
