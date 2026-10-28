# PIR configuration catalogue

The catalogue records Lift-RS code configurations for parameter planning. A partition
plan references configuration IDs and specifies application submatrices and
query-sharing groups; it is separate from this catalogue.

[configs_catalogue.json](configs_catalogue.json) contains all agreed `s=8..20`, `m=2`
configurations. Every entry retains Monte Carlo evidence from the
[rate experiment](../experiments/liftrs-rates/report.md), with sample count,
confidence interval, and reproducible evidence. The four matching paper estimates
from [Table 1](../paper.pdf#page=64) and [Table 3](../paper.pdf#page=66) remain as
supporting sources. The security profile comes from the separate
[bounded-sample calculation](../experiments/bounded-sample-parameters/report.md).
The s=8,m=2,t=8 entry now has exact K=3121, polynomial 0x11d and standalone
bounded orbit encoder support, backed by the [encoder comparison](../experiments/encoder-comparison/report.md).
Its historical Monte Carlo point estimate was K=3117 and remains separate evidence.
The s=9,m=2,t=7 entry now has exact K=17200, polynomial 0x211 and the same
standalone bounded support, backed by the [table-free checkpoint](../experiments/tablefree-encoding/report.md).
Its historical Monte Carlo point estimate K=17118 remains separate evidence.
The s=10,m=2,t=6 entry now has exact K=103556, polynomial 0x409 and standalone
bounded support, backed by the [CPU scaling checkpoint](../experiments/encoder-scaling/report.md).
Its historical Monte Carlo point estimate K=103952 remains separate evidence.
The s=11/12 entries now have exact K=455434/1965739 and polynomials 0x805/0x1053,
backed by [in-memory cyclic measurements](../experiments/cyclic-scaling/report.md).
Their historical Monte Carlo points K=453002/1964713 remain separate evidence.
Those runs do not establish large-cache JSON loading or paper-encoder performance.
The other 8 entries retain estimated capacities and candidate encoding status.
“Supported” identifies the non-cryptographic backend, not integrated application
encoding, grouped storage or production security. Historical cost-grid snapshots
retain their original estimates so objective comparisons use identical inputs.

The [tiny fixture](testdata/reference.json) has verified `K=5`, `N=16` over GF(4),
backed by the [reference encoder tests](../encoding/liftrs/reference_test.go).
It makes no security claim and is not part of the target security profile.

## Format

Documents use [JSON Schema draft 2020-12](schema.json), `version: 1`, and
`construction: "curve-lifted-rs-qminus1"`: restrictions to degree-at-most-`t` curves
have reduced degree at most `q-2`. Field elements use the agreed polynomial basis.
All links inside JSON are relative to the repository root (or durable URLs).

| Field | Meaning |
|---|---|
| `id` | Stable, descriptive configuration name, unique within the catalogue |
| `encoding_status` | `candidate` or `supported`; separate from capacity evidence |
| `code` | `s`, `m`, `t`, and the field polynomial's coefficient bitmask when selected |
| `capacity` | `status`, encoded length `n`, and message dimension `k` and code `rate` when known |
| `query` | `q-1` curve positions, fixed dummy count, total `length`, and dummy rule |
| `security` | Conjectured computational bits, base-two sample-budget exponent, assumptions, and evidence |
| `encoder` | Implementation, code-preprocessing method, and supporting evidence, when available |
| `measurements` | Optional resource observations with explicit stage, units, and evidence |

Unknown capacity omits `k` and `rate`. Estimated capacity records a rounded integer
point estimate `k` and `rate=k/n`. For paper estimates, the evidence states the
reported rate and conversion `k=round(reported_rate*n)`. The current catalogue uses
`k=round(n*good/samples)` from the new experiment as its point estimate. Monte Carlo evidence additionally records the
sampling method, sample count, confidence level, and lower/upper rate bounds. These
estimates can guide planning but cannot establish encoding capacity. Verified
capacity needs encoding or reference-test evidence. Declared encoding support also
requires a selected field polynomial and an encoder implementation.

Evidence kinds are `paper`, `analytical`, `monte_carlo`, `encoding`, `reference_test`,
and `measurement`. Every claim links to its source through `reference`; analytical
calculations, experiments, and measurements also require `script` and `results`.
New experiments must follow the [experiment conventions](../experiments/README.md).
Use the report to record revision, environment, methodology, and limitations.

`rate` means `k/n`, excluding application padding, metadata, and byte alignment.
Measured `stored_bytes` is separate. Resource quantities are `elapsed_seconds`,
`peak_bytes`, `artifact_bytes`, and `stored_bytes`. Their stage distinguishes reusable
`code_preprocessing` from per-database `database_encoding`. These observations do
not influence partition selection.

Version 1 accepts `s=2..20`, `m=2..3`, including tiny fixtures. Integer fields are
unsigned JSON numbers; use an integer-preserving parser. The target `m=2` dimensions
fit exactly in binary64, but some permitted `m=3` dimensions and capacities may not.
Omit absent optional fields; explicit `null` is rejected.

## Go loading

Import `github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig`:

```go
file, err := os.Open("pirconfig/configs_catalogue.json")
if err != nil {
    return err
}
defer file.Close()
configs, err := pirconfig.Load(file)
if err != nil {
    return err
}
// Use estimated capacities for provisional plans; verify capacity before encoding.
```

The [runnable example](example_test.go) loads the PIR configuration catalogue. `Load` rejects
unknown or duplicate keys, missing required fields, nulls, and trailing JSON. It
also checks unique IDs, `n=2^(s*m)`, `rate=k/n`, query lengths, polynomial
irreducibility, uncertainty intervals, and evidence requirements. `Validate` checks
in-memory documents after construction or modification. Returned data belongs to
the caller; validate it again after changes.

Loading never runs preprocessing or encoding, executes scripts, or opens evidence
links. Validation checks consistency, not the truth of evidence or security claims.
The application encoder must check the referenced backend supports the chosen code.
For untrusted input, bound the input size before calling `Load`.

## Verification and remaining work

```sh
go test ./catalogue
go test -race ./catalogue
```

[Tests](catalogue_test.go) compare the target profile with its documented CSV and
verify the tiny fixture against the actual encoder. They exercise malformed JSON,
arithmetic errors, unsupported capacity claims, Monte Carlo metadata, and resource
measurements. Synthetic test estimates are not experimental results.

Optional schema checks use Python and `jsonschema==4.25.1`:

```sh
python3 -m venv /tmp/sk-depir-pirconfig-schema-env
/tmp/sk-depir-pirconfig-schema-env/bin/pip install jsonschema==4.25.1
/tmp/sk-depir-pirconfig-schema-env/bin/python pirconfig/schema_test.py
```

Schema validation checks document structure and status-dependent requirements;
the Go loader additionally checks mathematical consistency. Neither is an encoder
registry. The [plan evaluator](../planner/README.md) consumes this catalogue for
capacity checks, cost calculations, and copied provenance. The [greedy planner](../planner/README.md#greedy-planning)
selects row-block plans from known capacities, with application logical-width search
and optional ratio-bounded full-download suffixes. Target-size encoding remains to be implemented. See the [planner design](../docs/parameter-planning.md).

The [bounded tradeoff search](../planner/README.md#storage-and-communication-tradeoffs)
uses the same catalogue evidence and returns storage/communication alternatives with
approximation flags. Communication rate is requested record bits divided by total
upload plus download per lookup.
