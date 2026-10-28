package liftrs

import (
	"encoding/json"
	"fmt"
)

// Encoder artifacts include all method-specific polynomial constants and cached
// solver reciprocals. Unlike the discovery artifact, loading them never rebuilds
// product trees or performs polynomial preparation. Field and transform metadata are rebuilt
// as bounded process-local runtime state; benchmark that load cost separately.
type orbitEncoderArtifact struct {
	Spectrum []uint32           `json:"generator_spectrum,omitempty"`
	Version  int                `json:"version"`
	Method   string             `json:"method"`
	Basis    json.RawMessage    `json:"basis"`
	P        []uint32           `json:"p"`
	G        []uint32           `json:"g,omitempty"`
	Weights  []uint32           `json:"weights,omitempty"`
	Tree     *orbitTreeArtifact `json:"tree,omitempty"`
}
type orbitTreeArtifact struct {
	Polynomial  []uint32 `json:"polynomial"`
	Reciprocal  []uint32 `json:"reciprocal,omitempty"`
	Root        uint32   `json:"root,omitempty"`
	Left, Right *orbitTreeArtifact
}

func saveOrbitTree(t *productTree) *orbitTreeArtifact {
	if t == nil {
		return nil
	}
	return &orbitTreeArtifact{t.polynomial, t.reciprocal, t.root, saveOrbitTree(t.left), saveOrbitTree(t.right)}
}
func loadOrbitTree(a *orbitTreeArtifact, n, reciprocal int, q uint32) (*productTree, error) {
	if a == nil || len(a.Polynomial) != n+1 || len(a.Reciprocal) != reciprocal || a.Root >= q {
		return nil, fmt.Errorf("invalid solver tree shape")
	}
	if a.Polynomial[n] != 1 {
		return nil, fmt.Errorf("nonmonic solver tree")
	}
	for _, v := range [][]uint32{a.Polynomial, a.Reciprocal} {
		for _, x := range v {
			if x >= q {
				return nil, fmt.Errorf("solver coefficient outside field")
			}
		}
	}
	t := &productTree{polynomial: a.Polynomial, reciprocal: a.Reciprocal, root: a.Root}
	if n == 1 {
		if a.Left != nil || a.Right != nil {
			return nil, fmt.Errorf("invalid solver leaf")
		}
		return t, nil
	}
	h := n / 2
	var err error
	t.left, err = loadOrbitTree(a.Left, h, n-h, q)
	if err != nil {
		return nil, err
	}
	t.right, err = loadOrbitTree(a.Right, n-h, h, q)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// MarshalBinary serializes the complete public paper solver as versioned JSON,
// including product-tree reciprocals and denominator inverses. It contains no
// database data or cryptographic keys and may be distributed with code provenance.
func (e *PaperEncoder) MarshalBinary() ([]byte, error) {
	basis, err := e.basis.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return json.Marshal(orbitEncoderArtifact{Version: 1, Method: "paper", Basis: basis, P: e.p, Weights: e.weights, Tree: saveOrbitTree(e.tree)})
}

// MarshalBinary serializes the complete public cyclic encoder as versioned JSON.
// Both multiplication constants P and G are stored, avoiding generator construction
// on load or for each database. Retain provenance when publishing this artifact.
func (e *CyclicEncoder) MarshalBinary() ([]byte, error) {
	basis, err := e.basis.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return json.Marshal(orbitEncoderArtifact{Version: 1, Method: "cyclic", Basis: basis, P: e.p, G: e.g, Spectrum: e.spectrum})
}
func loadEncoderArtifact(data []byte, method string) (*OrbitBasis, orbitEncoderArtifact, error) {
	var a orbitEncoderArtifact
	if len(data) > 64<<20 {
		return nil, a, fmt.Errorf("encoder artifact exceeds 64 MiB")
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, a, err
	}
	if a.Version != 1 || a.Method != method {
		return nil, a, fmt.Errorf("unsupported encoder artifact")
	}
	b, err := LoadOrbit(a.Basis)
	if err != nil {
		return nil, a, err
	}
	if len(a.P) != b.Dimension()+1 || a.P[0] != 1 {
		return nil, a, fmt.Errorf("invalid check polynomial")
	}
	limit := b.base.q
	if method == "paper" {
		limit = b.extension.q
	}
	for _, x := range a.P {
		if x >= limit {
			return nil, a, fmt.Errorf("check coefficient outside field")
		}
	}
	return b, a, nil
}

// LoadPaperEncoder loads trusted public preprocessing data, checking bounded
// shapes and field ranges. It does not independently prove the artifact encodes
// the intended code. No product-tree or polynomial preparation is performed.
func LoadPaperEncoder(data []byte) (*PaperEncoder, error) {
	b, a, err := loadEncoderArtifact(data, "paper")
	if err != nil {
		return nil, err
	}
	if len(a.Weights) != b.Dimension() || len(a.G) != 0 {
		return nil, fmt.Errorf("invalid solver weights")
	}
	for _, x := range a.Weights {
		if x == 0 || x >= b.extension.q {
			return nil, fmt.Errorf("invalid denominator inverse")
		}
	}
	tree, err := loadOrbitTree(a.Tree, b.Dimension(), 0, b.extension.q)
	if err != nil {
		return nil, err
	}
	return &PaperEncoder{b, tree, a.P, a.Weights}, nil
}

// LoadCyclicEncoder loads trusted public polynomial constants. Loading rebuilds
// field/FFT runtime tables but performs no generator construction or division.
// Structural checks do not replace independent verification of public artifacts.
func LoadCyclicEncoder(data []byte) (*CyclicEncoder, error) {
	b, a, err := loadEncoderArtifact(data, "cyclic")
	if err != nil {
		return nil, err
	}
	if len(a.G) != b.Length()-b.Dimension() || a.G[0] != 1 || a.Tree != nil || len(a.Weights) != 0 {
		return nil, fmt.Errorf("invalid cyclic generator shape")
	}
	for _, x := range a.G {
		if x >= b.base.q {
			return nil, fmt.Errorf("generator coefficient outside field")
		}
	}
	if len(a.Spectrum) != 0 && len(a.Spectrum) != b.Length() {
		return nil, fmt.Errorf("invalid generator spectrum size")
	}
	for _, x := range a.Spectrum {
		if x >= b.extension.q {
			return nil, fmt.Errorf("spectrum coefficient outside extension field")
		}
	}
	return &CyclicEncoder{basis: b, p: a.P, g: a.G, spectrum: a.Spectrum}, nil
}
