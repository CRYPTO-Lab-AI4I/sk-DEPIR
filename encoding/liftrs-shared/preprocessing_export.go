package liftrs

import (
	"bufio"
	"fmt"
	"io"
)

// PreparedPolynomials owns immutable, fully verified base-field P and G without
// an N-element spectrum. It holds no support or discovery/recovery frontier.
// Root checkpoint data aliases P, not another allocation. Results survive the
// session that constructed them. Do not expose private slices as mutable aliases.
type PreparedPolynomials struct {
	header       ScalingCodeHeader
	field        wideOrbitField
	p, generator []uint32
	root         *RootCheckpoint
	parameters   SupportParameters
	supportSHA   string
}

// Header describes the canonical code file that WriteCode will emit, including
// the spectrum not yet computed. It does not describe resident spectrum storage.
func (p *PreparedPolynomials) Header() ScalingCodeHeader {
	if p == nil {
		return ScalingCodeHeader{}
	}
	return p.header
}

// WriteGenerator writes verified G as canonical little-endian base-field slots.
// The writer must not retain mutable aliases; output may be partial on error.
func (p *PreparedPolynomials) WriteGenerator(destination io.Writer) error {
	if p == nil || destination == nil {
		return fmt.Errorf("verified polynomials/writer required")
	}
	return writeScalingCodeValues(destination, 4, uint64(1)<<uint(p.header.Spec.FieldDegree), p.generator, nil)
}

// PreparePolynomials performs the same root, reciprocal and complete P*G
// verification as Prepare, without allocating a spectrum. It preserves caller
// support/checkpoints; callers may release their own references after success.
// The resident Prepare entry point remains the default complete-code path.
func (p *CodePreprocessor) PreparePolynomials(support *CodeSupport, reuse PreparationReuse) (result *PreparedPolynomials, err error) {
	if p == nil {
		return nil, fmt.Errorf("nil preprocessor")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, fmt.Errorf("preprocessor closed")
	}
	defer recoverPreparation(&err)
	result, err = p.preparePolynomials(support, reuse, newPreparationReporter(p.options.Progress))
	if err == nil {
		err = p.savePreparedRecovery(result)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

type spectrumExporter interface {
	SpectrumTo([]uint32, int, io.Writer, int) error
}

// WriteCode explicitly computes and exports a canonical header/P/spectrum using
// a resident GPU executor and one bounded host tile. It does not retain a full
// host spectrum or add transforms. tileBytes is8 bytes..64MiB, rounded down to8 bytes.
// Chunk transfers/writes may cost time; no automatic fallback or mode selection
// occurs. The polynomial specification must match this session's field/code.
// Call WriteGenerator separately to preserve reproducible G. Output is partial
// on any error; callers own hashing, atomic publication and bounded verification.
// The writer runs under the session lock and must not reenter this session.
func (p *CodePreprocessor) WriteCode(destination io.Writer, polynomials *PreparedPolynomials, tileBytes int) (err error) {
	if p == nil || polynomials == nil || destination == nil || tileBytes < 8 || tileBytes > 64<<20 {
		return fmt.Errorf("streamed code session/polynomials/writer/tile required")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	defer recoverPreparation(&err)
	h := polynomials.header
	if p.closed || h.Spec != p.header.Spec || h.N != p.header.N || h.Polynomial != p.header.Polynomial || h.Delta != p.header.Delta || h.Omega != p.header.Omega {
		return fmt.Errorf("streamed code session/parameters differ")
	}
	exporter, ok := p.backend.(spectrumExporter)
	if !ok {
		return fmt.Errorf("streamed code export requires the resident CUDA backend")
	}
	p.backend.ReleaseProducts()
	w := bufio.NewWriterSize(destination, 64<<10)
	if err = writeScalingCodeHeader(w, h); err != nil {
		return err
	}
	if err = writeScalingCodeValues(w, 4, uint64(1)<<uint(h.Spec.FieldDegree), polynomials.p, nil); err != nil {
		return err
	}
	reporter := newPreparationReporter(p.options.Progress)
	if err = reporter.run("generator-spectrum-stream", func() error { return exporter.SpectrumTo(polynomials.generator, int(h.N), w, tileBytes) }); err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}
	return p.savePreparedRecovery(polynomials)
}
