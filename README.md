# deb822

A Go [SerDes](https://en.wikipedia.org/wiki/SerDes) library for Debian 
[deb822](https://www.debian.org/doc/debian-policy/ch-controlfields.html) encoding.

Supported document types: binary package stanzas (`types.Package`), repository
Release/InRelease files (`types.Release`), Sources index entries (`types.Source`),
source control files (`types.Dsc`) and upload control files (`types.Changes`).
OpenPGP clearsigned input is verified transparently when a keyring is supplied.

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
