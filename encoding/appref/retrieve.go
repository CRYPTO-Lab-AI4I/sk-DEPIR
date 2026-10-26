package appref

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
)

// ReadAt is an ordinary byte-range storage operation for a coordinate-identified
// PIR database. Like io.ReaderAt, it must return the number of bytes read and an
// error on short reads. It receives no group membership, query state, or keys.
// Remote adapters must honor ctx. The reference calls it once per occurrence,
// including repeated addresses, and never requires server-side computation.
type ReadAt func(ctx context.Context, id Coordinates, p []byte, offset int64) (int, error)

type query struct {
	positions []int // public storage addresses
	real      []int // private indices into positions, retaining curve multiplicity
}

// Retrieve reconstructs one application record using every fixed group and
// every PIR database. Irrelevant databases still receive their complete reads.
// Query generation uses crypto/rand; shared generation never removes repeated
// uploads. Any validation or read failure returns nil rather than a partial record.
// Correctness on these tiny fixtures does not establish a security level.
func (c *Client) Retrieve(ctx context.Context, row uint64, read ReadAt) ([]byte, error) {
	return c.retrieve(ctx, row, read, rand.Reader)
}

// The entropy argument exists only for explicit deterministic/error test fixtures.
func (c *Client) retrieve(ctx context.Context, row uint64, read ReadAt, entropy io.Reader) ([]byte, error) {
	if c == nil || len(c.databases) == 0 {
		return nil, fmt.Errorf("uninitialized application reference client")
	}
	if ctx == nil || read == nil {
		return nil, fmt.Errorf("context and storage reader must be non-nil")
	}
	if row >= c.plan.Rows {
		return nil, fmt.Errorf("application row %d is outside [0,%d)", row, c.plan.Rows)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Prepare all public requests before any I/O, then finish every read before
	// relevance-dependent unmasking/reconstruction. This keeps that private work
	// out of the sequence of storage calls (not a constant-time security claim).
	queries := make([]query, len(c.groups))
	for index, g := range c.groups {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		local := 0 // Valid cover index when no member contains the application row.
		for _, member := range g.members {
			p := c.databases[member].partition
			if contains(p.Coordinates, row) {
				local = int(row - p.RowStart)
				break
			}
		}
		var err error
		queries[index], err = makeQuery(g, local, entropy)
		if err != nil {
			return nil, fmt.Errorf("generate query: %w", err)
		}
	}
	responses := make([][]uint32, len(c.databases))
	for index, g := range c.groups {
		q := queries[index]
		for _, member := range g.members {
			d := c.databases[member]
			values := make([]uint32, len(q.positions))
			for i, position := range q.positions {
				var err error
				values[i], err = readElement(ctx, read, d.partition.Coordinates, position, d.partition.Code.FieldDegree)
				if err != nil {
					return nil, fmt.Errorf("read database %+v: %w", d.partition.Coordinates, err)
				}
			}
			responses[member] = values
		}
	}
	record := make([]byte, (int(c.plan.RecordBits)+7)/8)
	for member, d := range c.databases {
		if !contains(d.partition.Coordinates, row) {
			continue
		}
		q := queries[int(d.partition.Group)]
		positions := make([]int, len(q.real))
		values := make([]uint32, len(q.real))
		for i, index := range q.real {
			positions[i] = q.positions[index]
			values[i] = responses[member][index]
		}
		plain, err := d.masker.Unmask(positions, values)
		if err != nil {
			return nil, err
		}
		var value uint32
		for _, part := range plain {
			value ^= part
		}
		for col := d.partition.ColumnStart; col < d.partition.ColumnEnd; col++ {
			bit := (value >> uint(col-d.partition.ColumnStart)) & 1
			record[int(col/8)] |= byte(bit) << uint(col%8)
		}
	}
	return record, nil
}

func contains(id Coordinates, row uint64) bool { return row >= id.RowStart && row < id.RowEnd }

// Each coordinate is a uniformly sampled degree-at-most-t polynomial with its
// constant coefficient fixed to the target coordinate. Evaluating at z!=0 gives
// a punctured curve through the target at zero. Locality q-1 permits decoding by
// summing its field values, including repeated curve-point occurrences.
func makeQuery(g group, local int, entropy io.Reader) (query, error) {
	spec := g.code.Spec()
	q := 1 << uint(spec.FieldDegree)
	target := g.code.SystematicPositions()[local]
	coefficients := make([][]uint32, spec.Dimension)
	for coordinate := range coefficients {
		coefficients[coordinate] = make([]uint32, spec.CurveDegree+1)
		coefficients[coordinate][0] = uint32(target % q)
		target /= q
		for power := 1; power <= spec.CurveDegree; power++ {
			value, err := randomPowerOfTwo(entropy, q)
			if err != nil {
				return query{}, err
			}
			coefficients[coordinate][power] = uint32(value)
		}
	}
	curve := make([]int, q-1)
	for z := 1; z < q; z++ {
		position, stride := 0, 1
		for _, polynomial := range coefficients {
			value := polynomial[len(polynomial)-1]
			for power := len(polynomial) - 2; power >= 0; power-- {
				value = multiply(value, uint32(z), uint32(q), g.code.Polynomial()) ^ polynomial[power]
			}
			position += int(value) * stride
			stride *= q
		}
		curve[z-1] = position
	}
	positions, err := g.permutation.Map(curve)
	if err != nil {
		return query{}, err
	}
	type entry struct {
		position int
		real     bool
	}
	entries := make([]entry, 0, len(positions)+DummyPositions)
	for _, position := range positions {
		entries = append(entries, entry{position, true})
	}
	for i := 0; i < DummyPositions; i++ {
		position, err := randomPowerOfTwo(entropy, g.code.Length())
		if err != nil {
			return query{}, err
		}
		entries = append(entries, entry{position, false})
	}
	for i := len(entries) - 1; i > 0; i-- {
		j, err := rand.Int(entropy, big.NewInt(int64(i+1)))
		if err != nil {
			return query{}, err
		}
		entries[i], entries[int(j.Int64())] = entries[int(j.Int64())], entries[i]
	}
	out := query{positions: make([]int, len(entries))}
	for i, e := range entries {
		out.positions[i] = e.position
		if e.real {
			out.real = append(out.real, i)
		}
	}
	return out, nil
}

// q and N are powers of two <=256 in this backend, so byte masking is uniform.
func randomPowerOfTwo(entropy io.Reader, bound int) (int, error) {
	var b [1]byte
	if _, err := io.ReadFull(entropy, b[:]); err != nil {
		return 0, err
	}
	return int(b[0]) & (bound - 1), nil
}

func multiply(a, b, q, polynomial uint32) uint32 {
	var result uint32
	for b != 0 {
		if b&1 != 0 {
			result ^= a
		}
		b >>= 1
		a <<= 1
		if a&q != 0 {
			a ^= polynomial
		}
	}
	return result
}

func readElement(ctx context.Context, read ReadAt, id Coordinates, position, width int) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	start := position * width
	begin, end := start/8, (start+width+7)/8
	var buffer [2]byte // Fixture widths <=4 span at most two bytes.
	p := buffer[:end-begin]
	n, err := read(ctx, id, p, int64(begin))
	if n < 0 || n > len(p) {
		return 0, fmt.Errorf("storage reader returned invalid byte count %d", n)
	}
	if n != len(p) {
		if err == nil {
			err = io.ErrUnexpectedEOF
		}
		return 0, fmt.Errorf("short read: got %d bytes, want %d: %w", n, len(p), err)
	}
	if err != nil && err != io.EOF {
		return 0, err
	}
	var value uint32
	for bit := 0; bit < width; bit++ {
		i := start%8 + bit
		value |= uint32((p[i/8]>>uint(i%8))&1) << uint(bit)
	}
	return value, nil
}
