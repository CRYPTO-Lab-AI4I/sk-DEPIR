# Experimental scaling code data

LoadScalingCode and VerifyScalingCode are separate code-data APIs for the s17
port. They consume trusted reusable mathematical preprocessing, not plaintext or
encoded application databases. This format is experimental; neither a readable
header nor a matching hash proves polynomial identities, verified K, code
membership, cryptographic security or full s17 encoder support.

[Implementation](../encoding/liftrs-shared/scaling_code.go),
[API example](../encoding/liftrs-shared/scaling_code_example_test.go),
[validation experiment](../experiments/s17-code-data/report.md).

## Payload and size selection

All forms describe m=2, N=2^(2s), with K+1 base-field P coefficients followed by
N extension-field generator-spectrum values. Integers are unsigned little-endian.
P coefficients always use4-byte slots. Width depends on s; no resident copy of a
legacy spectrum is widened unnecessarily.

| s | Canonical magic | Header bytes | K/N count bytes each | P slot bytes | Spectrum slot bytes |
|---|---|---|---|---|---|
| 2–15 | LRSprobe1 | 25 | 4 | 4 | 4 |
| 16 | LRSprobe2 | 33 | 8 | 4 | 4 |
| 17–20 | LRSprobe3 | 53 | 8 | 4 | 8 |

The first two forms retain the existing encoding, including their implicit
conventional polynomial, deterministic tower delta and first primitive generator.
The loader also accepts legacy LRSprobe2 at smaller s; writing normalizes such
counts to the canonical width above. Canonical old files round-trip byte-for-byte.
The old LoadScalingOnline/encoders still reject the new magic and retain s≤16.
Existing files are never overwritten or migrated automatically.

LRSprobe3 has these byte offsets:

| Offset | Field |
|---|---|
| 0 | 9-byte magic LRSprobe3 |
| 9 | s, uint32 |
| 13 | curve degree t, uint32 |
| 17 | K, uint64 |
| 25 | N, uint64 |
| 33 | base polynomial identifier, uint32 |
| 37 | tower delta, uint32 |
| 41 | packed primitive generator omega, uint64 |
| 49 | four bytes: count width8, P width4, spectrum width8, reserved0 |
| 53 | K+1 P values, then N spectrum values |

The new identifiers must match this implementation's conventional field/tower
and generator ordering; an incompatible representation is rejected, never silently
reinterpreted. This is not support for an arbitrary field isomorphism. Base values
are less than2^s; extension values/coordinates less than2^(2s), packed as a|(b<<s).
P[0] must be1. Shape checks require1≤K≤N/2; they do not guarantee the current
2L≤N encoder workspace fits, with L=nextpow2(2K−1). In particular s20 may need
an expanded prefix layout even if its code-data representation is readable.

## Loading, verification and persistence

ReadScalingCodeHeader reads at most53 bytes without domain arrays. It validates
header representation only and returns unverified metadata. Offset arithmetic
and systematic indices/exponents use uint64. Counts and dimensions are checked
before host-length conversion or allocation.

LoadScalingCode requires a SHA-256 from an independently trusted manifest and
positive ScalingCodeLimits.MaxElements/MaxPayloadBytes. Principal-array storage
is exactly4(K+1)+wN bytes, where w is4/8 above. Caller RAM budgets must allow
bounded64KiB I/O buffers, field metadata, runtime/allocator and headroom beyond
this cap. A header claiming too much is rejected before principal allocation.
Loading checks every value and exact EOF while computing the digest in one pass.
The returned data holds private arrays and exposes header, resident-byte count
and zero-based systematic positions; it is not a larger online encoder plan.

VerifyScalingCode performs the same shape/range/P-constant/EOF/digest checks with
bounded I/O memory and no resident payload arrays. Limits cap its logical payload
extent, rather than allocated arrays. It is a standalone validation option, not
an automatic second pass: calling Verify then Load would read/hash the data twice.
No full s17 file has yet been prepared or verified by this prototype.

WriteScalingCode streams canonical slot/count widths without changing values.
It checks metadata/ranges and owns neither a destination path nor publication.
The caller must persist atomically and save the new digest/provenance; errors
can leave a partial destination. It does not independently verify code identities.
New s17 files require independently verified preparation before encoding admission.
Full preprocessing, native code-data integration and encoding remain separate
[plan gates](encoding-scaling.md#current-execution-plan-for-s17).
