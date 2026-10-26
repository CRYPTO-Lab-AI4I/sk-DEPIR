# Small cryptographic encoding reference

This package provides a bounded in-memory correctness reference for the
cryptographic encoding stage: permute a codeword and XOR each field value with
a mask bound to its storage position. It reuses the existing AES helpers.
Its behavior can be compared against future implementations; its tables, slices,
I/O strategy, and API do not prescribe the architecture for large databases.

For a codeword `c` and codeword-to-storage permutation `pi`:

`stored[pi(i)] = c[i] XOR mask_k(pi(i))`.

The client reads the requested storage positions and removes their masks before
combining field elements for code decoding. Storage only performs ordinary reads.
[example_test.go](example_test.go) shows preprocessing, code encoding, fresh keys,
cryptographic encoding, and unmasking at systematic positions. That example's
direct reads illustrate the boundary, not a private PIR query; the
[retrieval test](../../pir/cryptoref_test.go) uses curve queries and dummies.

## Operations and key scopes

- `GenerateKey()` obtains a fresh 32-byte key using
  [`crypto/rand`](https://pkg.go.dev/crypto/rand). Call it independently for
  each query-sharing group's permutation and each PIR database encoding's masks.
- `NewPermutation(length, key)` snapshots the existing
  [AES permutation](../../utils/permutator.go) with an empty tweak and checks
  bijectivity across its entire tiny domain. Its `Map` method converts codeword
  positions to storage positions, retaining order and repetitions.
- `NewMasker(fieldDegree, length, key)` holds one encoding's mask key.
  `Encode(codeword, permutation)` scatters and XOR-masks the values.
  `Unmask(storagePositions, responses)` returns unmasked values in read order;
  the client still performs code decoding.

PIR databases sharing a query use the same permutation but independent masking
keys. A re-encoding uses a fresh masking key even if its coordinate identifier
is unchanged. Constructors cannot detect reuse across calls; the caller owns
this key lifecycle. Keys and permutation maps are secret and are not identifiers
or public metadata. Unmasking does not authenticate storage contents.

## Exact reference behavior and limits

Lengths are 2 through 256, and field degrees are 2, 3, or 4. AES constructors
accept 16-, 24-, or 32-byte keys; generated keys are 32 bytes. The package copies
key state, does not modify inputs, and allocates separate outputs. Initialized
objects support concurrent reads and encoding of the same data. Do not reuse a
masker/key for another database or changed data.

Mask values reuse `Encryptor.EncryptPosition(0, position)` from
[encryptor.go](../../utils/encryptor.go). Its AES input is
`LE64(storage_position) || LE64(0)`. Interpret the encrypted block as a big-endian
integer and take its low `s` bits, equivalent to reduction modulo `2^s`.
Use XOR to add or remove this field mask. The helper's integer-modulus
`Encrypt`/`Decrypt` operations are not used by this reference stage.
Separate encoding keys distinguish databases and re-encodings; the zero in the
AES input is a fixed label, not a database identifier.

The existing permutation is custom prototype code. Checking its bijectivity and
testing masks establishes correctness on these fixtures, not production security.
The legacy file-based RM API and its integer-modulus masking remain separate.
Online curve generation still uses prototype randomness. Streaming, chunked I/O,
external-memory permutation, and the eventual public API require their own work;
they need only match the agreed encoding behavior, not this in-memory design.

## Verification

```sh
go test ./encoding/cryptoref
go test ./pir -run '^TestCryptographicReferenceLiftedRSRecovery$'
go test -race ./encoding/cryptoref
```

[Unit tests](reference_test.go) check entire permutation domains, masks against
direct AES evaluations, a fixed zero-block vector, XOR round trips for every
fixture field value, database/re-encoding key separation, repeated addresses,
invalid input, ownership, and concurrent calls. Individual masks can coincide in
a small field; separation tests compare full fixed-key mask vectors.

The [integration test](../../pir/cryptoref_test.go) retrieves every systematic
position of genuine GF(16) lifted codewords using the existing curve-query
sampler, 128 dummies, and plain in-memory reads. Two databases and a re-encoding
share the generated query, with a separate full upload for each database.
Deterministic degree-2 curves explicitly exercise repeated positions. See the
[encoder contract](../../docs/encoder-contract.md#cryptographic-encoding) for the
behavior this reference implements.
