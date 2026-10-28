# Lift-RS transform execution in Go

[transform.go](transform.go) contains the pure-Go additive-transform recursion and
fixed-constant multiplication evaluator. It borrows immutable tables prepared by
[liftrs-shared](../liftrs-shared/README.md). It performs no code discovery, public
artifact loading, full encoder orchestration, masking or application partitioning.
Scalar field arithmetic is shared through `liftrs-shared/field`.

Run the complete Go path through the shared API:

```sh
CGO_ENABLED=0 go test ./encoding/liftrs-shared ./encoding/liftrs-go
# Runnable complete encoding example:
go test ./encoding/liftrs-shared -run ExampleOrbitBasis -v
```

Use `CyclicEncoder.Encode` from `encoding/liftrs-shared`. The prepared plan is an
internal execution view; applications should use that shared API rather than
construct raw transform tables. Recursion, four-bit rows and buffer ordering are
preserved by the extraction; existing timing reports predate this reorganization.
See the [encoding overview](../README.md) for the C++ alternative and directory map.
