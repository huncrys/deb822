# deb822

A Go [SerDes](https://en.wikipedia.org/wiki/SerDes) library for Debian 
[deb822](https://www.debian.org/doc/debian-policy/ch-controlfields.html) encoding.

Supported document types: binary package stanzas (`types.Package`), repository
Release/InRelease files (`types.Release`) and their per component stubs
(`types.ComponentRelease`), Sources index entries (`types.Source`), source
control files (`types.Dsc`) and upload control files (`types.Changes`).
OpenPGP clearsigned input is verified transparently when a keyring is supplied.
The `contents` and `changelog` packages additionally cover the archive's
`Contents-*` indices and Debian changelogs, which are not deb822 documents.

## Struct tags

Stanza serialization is driven by the dedicated `debian:` struct tag:

```go
type Example struct {
    Name     string          `debian:"Package" json:"Package"`
    Depends  dependency.Dependency `debian:"Depends,omitempty" json:"Depends,omitzero"`
    Ignored  string          `debian:"-" json:"-"`
}
```

- `debian:"Field-Name[,omitempty][,inline]"`; `debian:"-"` skips the field.
- If no `debian:` tag is present, the name part of the `json:` tag is used,
  then the Go field name - existing json-tagged structs keep working unmigrated.
- Field names match case-insensitively on decode (Debian Policy 5.1).
- `json:` tags are only used by `encoding/json`; `json.Marshal` of the built-in
  types no longer emits empty-string values (they carry `omitzero`).

## Strict mode

Parsing is lenient by default (unchanged from earlier releases). Debian Policy
5.1 validation is opt-in:

```go
dec, err := deb822.NewDecoder(r, keyring, deb822.WithStrict())
```

- `WithStrict()` enforces field-name syntax, rejects duplicate fields
  (case-insensitively) and rejects comment lines.
- `WithComments(allow bool)` overrides the comment policy independently, so
  `WithStrict(), WithComments(true)` suits `debian/control`-style files.
- Violations wrap the sentinel errors `ErrInvalidFieldName`,
  `ErrDuplicateField`, `ErrCommentNotAllowed` and `ErrUnexpectedContinuation`
  for use with `errors.Is`.

## Contents indices

`contents` reads and writes the `Contents-$arch` / `Contents-source` indices.
These are not deb822 documents - they are a flat two-column table - but they are
part of the same archive layout, and the `Release` file this library parses
carries their hashes.

```go
r := contents.NewReader(gzipReader)
for {
    entry, err := r.Read()
    if errors.Is(err, io.EOF) {
        break
    } else if err != nil {
        return err
    }
    // entry.Path, entry.Packages ([[area/]section/]package)
}
```

- Columns are split on the *last* whitespace run of the line: the separator has
  no fixed width (dak pads to 55 columns with spaces, apt-ftparchive pads with
  tabs) and real archives ship paths containing spaces.
- `NewWriter` emits dak's binary layout by default; `WithPadding(0)` plus
  `WithTabSeparator()` gives the `Contents-source` layout. Entries that could
  not be read back unambiguously are rejected with `ErrInvalidPath` /
  `ErrInvalidPackageList`.
- The legacy prose header terminated by a `FILE LOCATION` line is skipped on
  read and exposed via `Reader.Header()`; it is never written.
- `ParseQualifiedName` splits the second column, tolerating names with and
  without an area prefix.
- Compression is the caller's business; both ends take plain streams.

## Changelogs

`changelog` reads and writes Debian changelogs: `debian/changelog` in a source
package, and the `changelog.Debian.gz` a binary package ships in
`/usr/share/doc`. It is not a deb822 document either, but it is what
`dpkg-genchanges` and `dpkg-source` derive the `.changes` and `.dsc` this
library already models from.

```go
r := changelog.NewReader(gzipReader)
entries, err := r.ReadAll() // newest first
if errors.Is(err, changelog.ErrNotDebianFormat) {
    // the package ships an upstream changelog instead; treat it as opaque text
}
```

- `Entry.Changes` holds the body *verbatim*, blank lines and indentation
  included, and the writer plays it back untouched. Generators disagree about
  the layout - dpkg and `dch` put a blank line after the header, nFPM's
  goreleaser/chglog does not and indents continuations by three spaces - and
  none of it is normalized away.
- The header option list is exposed as `Urgency` plus an ordered `Options`
  slice, so the `binary-only=yes` of a binNMU survives a round trip;
  `Entry.BinaryOnly()` and `Entry.Option(key)` read it back.
- Ancient entries with no distribution (`hello (1.3-6); priority=LOW`) and
  uppercase urgencies parse as written.
- `ErrNotDebianFormat` distinguishes a file that never was a Debian changelog
  from a corrupt one, so an archive tool can publish an upstream changelog
  unparsed instead of rejecting the package. Free-form history past the last
  entry is exposed via `Reader.Trailing()` rather than failing the read.
- Trailer dates are always written as RFC1123 with a *numeric* zone, which is
  what dpkg requires - including for an entry built from a file mtime or
  `time.Now()`, whose zone carries a name. A changelog carrying the
  space-padded days of the nineties therefore reformats once and is a fixed
  point after that. Dates no layout
  accepts get one salvage pass, since `dpkg-parsechangelog` never parses that
  field and the archive shows it (bash ships `Thur, 19 June 1997`).
- Compression is the caller's business; both ends take plain streams.

## v0.11.0 changes

- New `types.ComponentRelease`: the per component, per architecture `Release`
  stub an archive publishes at
  `dists/<suite>/<component>/binary-<arch>/Release`. Field order follows
  Debian's own stubs.
- New `types.Package.DescriptionMD5Sum()`, returning the hex md5 apt publishes
  as `Description-md5`. The checksum is taken over the description in its
  on-wire form, so it cannot be computed from the decoded value: the decoder
  unfolds continuation lines and turns `" ."` back into a blank line, and apt
  hashes the folded text plus its terminating newline. An empty description
  yields `""`, matching apt, which records no checksum rather than the digest
  of a bare newline. Additive, nothing else changed.

## v0.10.1 changes

- `changelog`: the trailer date is now always written with a numeric zone.
  Building an entry from a file mtime or `time.Now()` emitted the zone *name*
  (`CEST` rather than `+0200`), which dpkg rejects. Files read through `Reader`
  were never affected.

## v0.10.0 changes

- New `changelog` package: a reader and writer for Debian changelogs (see
  above). Additive, nothing else changed.

## v0.9.0 changes

- New `contents` package: a reader and writer for the archive's `Contents-*`
  indices (see above). Additive, nothing else changed.

## v0.8.0 changes

- **Breaking:** `arch.Arch` now models dpkg's four-component architecture
  tuples (`abi-libc-os-cpu`, dpkg >= 1.18.11). The former `ABI` field actually
  held the libc component and is now named `Libc`; `ABI` holds the real ABI
  component (default `base`). `Parse("amd64")` yields
  `{ABI: "base", Libc: "gnu", OS: "linux", CPU: "amd64"}`. Wildcards left-pad
  with `any` as dpkg does, and `String()` emits dpkg-canonical short forms
  (`musl-linux-amd64`, never the ambiguous `musl-amd64`).
- Version parsing matches dpkg 1.22 behavior: empty revisions (`1.0-`), empty
  upstream versions (`1:-1`) and non-numeric epochs are rejected; epochs are
  capped at INT_MAX; upstream versions not starting with a digit are accepted
  (dpkg only warns).
- The legacy dependency operators `<` and `>` are accepted and normalized to
  `<=` and `>=`, matching dpkg.
- Dates in Release/Changes files accept RFC1123, numeric-zone, single-digit-day
  and weekday-less layouts on decode; encode is unchanged (RFC1123). Dates
  decoded from a numeric offset always re-encode with a numeric offset,
  regardless of the host timezone.
- `Files`/`Checksums-*` filenames keep interior spaces;
  `filehash.ChangesFileHash` covers the five-field `.changes` Files format.
- `types.Package` gained `Built-Using`, `Static-Built-Using`, `Package-Type`,
  `Task`, `Origin`, `Bugs`, `Build-Essential` and the debian-installer fields;
  `types.Release` gained `SHA512` and now spells
  `No-Support-for-Architecture-all` per the repository format spec.
- Input starting with a continuation line returns an error instead of
  panicking.

## Credits

This library was originally based on the work of the [go-debian](https://github.com/paultag/go-debian) 
library by Paul Tagliamonte but has since been heavily modified and extended.
