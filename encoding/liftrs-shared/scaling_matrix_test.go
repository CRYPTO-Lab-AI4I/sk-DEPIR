package liftrs

// This test-only disk prototype establishes exact general-basis Taylor/matrix
// semantics. Its scalar ReadAt/WriteAt transpose is intentionally NOT a high-
// throughput implementation. Counters are logical syscall bytes, not device I/O.
import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

type scalingDisk struct {
	file                 *os.File
	read, written, calls uint64
}

func (d *scalingDisk) get(i int) uint32 {
	var b [4]byte
	n, err := d.file.ReadAt(b[:], int64(i)*4)
	if err != nil || n != 4 {
		panic(fmt.Sprint("disk read: ", n, err))
	}
	d.read += 4
	d.calls++
	return binary.LittleEndian.Uint32(b[:])
}
func (d *scalingDisk) put(i int, x uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], x)
	n, err := d.file.WriteAt(b[:], int64(i)*4)
	if err != nil || n != 4 {
		panic(fmt.Sprint("disk write: ", n, err))
	}
	d.written += 4
	d.calls++
}

// genericMatrixPlan is not the canonical-only production constructor. It builds
// S_k over any ordered independent basis; Values still accepts canonical packed
// affine offsets, which are needed for the row cosets. No field-coordinate change.
func genericMatrixPlan(f orbitField, basis []uint32) *additiveTransform {
	levels := len(basis)
	width := bits.Len32(f.q) - 1
	a := &additiveTransform{f: f, terms: make([][]linearTerm, levels), values: make([][]uint32, levels), delta: make([]uint32, levels), inverse: make([]uint32, levels)}
	p := []uint32{1}
	eval := func(x uint32) uint32 {
		var y uint32
		for i, c := range p {
			y ^= f.mul(c, f.pow(x, 1<<uint(i)))
		}
		return y
	}
	for k := 0; k < levels; k++ {
		for i, c := range p {
			if c != 0 {
				a.terms[k] = append(a.terms[k], linearTerm{degree: 1 << uint(i), coefficient: c})
			}
		}
		a.values[k] = make([]uint32, width)
		for j := 0; j < width; j++ {
			a.values[k][j] = eval(1 << uint(j))
		}
		a.delta[k] = eval(basis[k])
		if a.delta[k] == 0 {
			panic("dependent matrix basis")
		}
		a.inverse[k] = f.inv(a.delta[k])
		next := make([]uint32, len(p)+1)
		for i, c := range p {
			next[i+1] ^= f.mul(c, c)
			next[i] ^= f.mul(a.delta[k], c)
		}
		p = next
	}
	return finishAdditiveTransform(a)
}
func matrixOffset(a *additiveTransform, k int, x uint32) uint32 {
	var y uint32
	for x != 0 {
		j := bits.TrailingZeros32(x)
		y ^= a.values[k][j]
		x &= x - 1
	}
	return y
}

// taylorDisk divides by h^(2^(v-1)) recursively. In characteristic two its
// coefficients are Frobenius powers. Forward stores the quotient in the upper
// half and remainder below; inverse reconstructs in ascending quotient order.
func taylorDisk(d *scalingDisk, f orbitField, terms []linearTerm, start, n, v int, inverse bool) {
	if v == 0 {
		return
	}
	half := n / 2
	k := v - 1
	if inverse {
		taylorDisk(d, f, terms, start, half, k, true)
		taylorDisk(d, f, terms, start+half, half, k, true)
	}
	if !inverse {
		for i := n - 1; i >= half; i-- {
			x := d.get(start + i)
			for _, term := range terms[:len(terms)-1] {
				degree := term.degree << uint(k)
				c := f.pow(term.coefficient, 1<<uint(k))
				at := start + i - half + degree
				d.put(at, d.get(at)^f.mul(c, x))
			}
		}
	} else {
		for j := 0; j < half; j++ {
			x := d.get(start + half + j)
			for _, term := range terms[:len(terms)-1] {
				degree := term.degree << uint(k)
				c := f.pow(term.coefficient, 1<<uint(k))
				at := start + j + degree
				d.put(at, d.get(at)^f.mul(c, x))
			}
		}
	}
	if !inverse {
		taylorDisk(d, f, terms, start, half, k, false)
		taylorDisk(d, f, terms, start+half, half, k, false)
	}
}

// matrixDisk uses one N-element file and a max(n1,n2) tile, plus small plans.
// The matrix consists of h-adic coefficient rows, h=S_split. Column evaluation
// uses h(high-basis); row evaluation uses the low basis with affine high offsets.
func matrixDisk(d *scalingDisk, f orbitField, dimension, split int, inverse bool) int {
	if split < 1 || split >= dimension {
		panic("invalid matrix split")
	}
	n1, n2 := 1<<uint(split), 1<<uint(dimension-split)
	low := make([]uint32, split)
	for j := range low {
		low[j] = 1 << uint(j)
	}
	// Need S_split in addition to S_0..S_(split-1).
	extra := append(append([]uint32(nil), low...), 1<<uint(split))
	hplan := genericMatrixPlan(f, extra)
	terms := hplan.terms[split]
	gamma := make([]uint32, dimension-split)
	for j := range gamma {
		gamma[j] = matrixOffset(hplan, split, 1<<uint(split+j))
	}
	columns, rows := genericMatrixPlan(f, gamma), genericMatrixPlan(f, low)
	// Column recursion offsets are encoded in the gamma basis, unlike row
	// offsets which are canonical packed coordinates. The executor's XOR indices
	// must therefore look up S_k(gamma_j), not S_k(2^j).
	for k := range columns.values {
		for j := range columns.values[k] {
			columns.values[k][j] = 0
			if j < len(gamma) {
				for _, term := range columns.terms[k] {
					columns.values[k][j] ^= f.mul(term.coefficient, f.pow(gamma[j], term.degree))
				}
			}
		}
	}

	tile := make([]uint32, maxIntScaling(n1, n2))
	doRows := func() {
		for j := 0; j < n2; j++ {
			buf := tile[:n1]
			for i := range buf {
				buf[i] = d.get(j*n1 + i)
			}
			rows.goPlan.Recurse(buf, split, uint32(j*n1), inverse)
			for i, x := range buf {
				d.put(j*n1+i, x)
			}
		}
	}
	doColumns := func() {
		for i := 0; i < n1; i++ {
			buf := tile[:n2]
			for j := range buf {
				buf[j] = d.get(j*n1 + i)
			}
			columns.goPlan.Recurse(buf, dimension-split, 0, inverse)
			for j, x := range buf {
				d.put(j*n1+i, x)
			}
		}
	}
	if !inverse {
		taylorDisk(d, f, terms, 0, n1*n2, dimension-split, false)
		doColumns()
		doRows()
	} else {
		doRows()
		doColumns()
		taylorDisk(d, f, terms, 0, n1*n2, dimension-split, true)
	}
	return len(tile) * 4
}

// routeDisk converts coefficient-orbit order into canonical grid order without
// an N-element RAM permutation. A high-coordinate bucket selects a contiguous
// output tile; each record stores its local uint32 offset and base value. Bounded
// fanout is enforced here by the small correctness fixture, not target-size claims.
func routeDisk(t *testing.T, b *OrbitBasis, y []uint32, tileElements int, dir string) ([]uint32, uint64) {
	t.Helper()
	n := b.Length()
	count := (n + tileElements - 1) / tileElements
	if count > 64 {
		t.Fatal("prototype bucket fanout exceeds 64")
	}
	files := make([]*os.File, count)
	defer func() {
		for _, f := range files {
			if f != nil {
				f.Close()
			}
		}
	}()
	for i := range files {
		f, err := os.Create(fmt.Sprintf("%s/bucket-%d", dir, i))
		if err != nil {
			t.Fatal(err)
		}
		files[i] = f
	}
	var traffic uint64
	point := b.omega
	parity := uint32(0)
	for _, x := range y[:n-1] {
		var record [8]byte
		bucket := int(point) / tileElements
		binary.LittleEndian.PutUint32(record[:4], point%uint32(tileElements))
		binary.LittleEndian.PutUint32(record[4:], x)
		if _, err := files[bucket].Write(record[:]); err != nil {
			t.Fatal(err)
		}
		traffic += 8
		parity ^= x
		point = b.extension.mul(point, b.omega)
	}
	// The returned full vector is a TEST sink for comparison. The routing loop
	// itself needs only one tile; real storage would append each completed tile.
	out := make([]uint32, n)
	for bucket, f := range files {
		if _, err := f.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		tile := make([]uint32, minInt(tileElements, n-bucket*tileElements))
		info, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}
		for at := int64(0); at < info.Size(); at += 8 {
			var record [8]byte
			if _, err := f.ReadAt(record[:], at); err != nil {
				t.Fatal(err)
			}
			traffic += 8
			offset := binary.LittleEndian.Uint32(record[:4])
			if int(offset) >= len(tile) {
				t.Fatal("route offset")
			}
			tile[offset] = binary.LittleEndian.Uint32(record[4:])
		}
		if bucket == 0 {
			tile[0] = parity
		}
		copy(out[bucket*tileElements:], tile)
		traffic += uint64(len(tile)) * 4
	}
	return out, traffic
}
func TestScalingMatrixDisk(t *testing.T) {
	for _, s := range []int{2, 3, 4} {
		polynomial, _ := BinaryPolynomial(s)
		_, f, _, _ := orbitFields(s, polynomial)
		for dimension := 2; dimension <= 2*s; dimension++ {
			n := 1 << uint(dimension)
			message := make([]uint32, n)
			random := rand.New(rand.NewSource(int64(701 + n)))
			for i := range message {
				message[i] = uint32(random.Intn(int(f.q)))
			}
			want := append([]uint32(nil), message...)
			newSparseTransform(f).applySubspace(want, false)
			for split := 1; split < dimension; split++ {
				file, err := os.CreateTemp(t.TempDir(), "matrix")
				if err != nil {
					t.Fatal(err)
				}
				d := &scalingDisk{file: file}
				for i, x := range message {
					d.put(i, x)
				}
				matrixDisk(d, f, dimension, split, false)
				got := make([]uint32, n)
				for i := range got {
					got[i] = d.get(i)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatal("matrix forward", s, dimension, split)
				}
				matrixDisk(d, f, dimension, split, true)
				for i, x := range message {
					if d.get(i) != x {
						t.Fatal("matrix inverse", s, dimension, split, i)
					}
				}
				file.Close()
			}
		}
	}
}
func TestScalingDiskConvolutionRouting(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(702))})
	if err != nil {
		t.Fatal(err)
	}
	e, err := b.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	message := make([]uint32, b.Dimension())
	for i := range message {
		message[i] = uint32((i*7 + 3) % 16)
	}
	// Exercise both partial-domain prefix and full-domain convolution with disk
	// forward/inverse, preserving ordinary convolution padding.
	multiply := func(a, z []uint32, limit int) []uint32 {
		n := 1
		for n < len(a)+len(z)-1 {
			n *= 2
		}
		dimension := bits.Len(uint(n)) - 1
		files := make([]*scalingDisk, 2)
		for j := range files {
			f, err := os.CreateTemp(t.TempDir(), "product")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			files[j] = &scalingDisk{file: f}
		}
		for j, values := range [][]uint32{a, z} {
			for i := 0; i < n; i++ {
				var x uint32
				if i < len(values) {
					x = values[i]
				}
				files[j].put(i, x)
			}
			matrixDisk(files[j], b.extension, dimension, dimension/2, false)
		}
		for i := 0; i < n; i++ {
			files[0].put(i, b.extension.mul(files[0].get(i), files[1].get(i)))
		}
		matrixDisk(files[0], b.extension, dimension, dimension/2, true)
		out := make([]uint32, limit)
		for i := range out {
			out[i] = files[0].get(i)
		}
		return out
	}
	h := multiply(e.p[:b.Dimension()], message, b.Dimension())
	y := multiply(e.g, h, b.Length()-1)
	// Final coefficient at N-1 is absent in ordinary product; pad for routing API.
	y = append(y, 0)
	want, err := e.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	for _, tile := range []int{4, 16, 64} {
		got, _ := routeDisk(t, b, y, tile, t.TempDir())
		if !reflect.DeepEqual(got, want) {
			t.Fatal("canonical routing", tile)
		}
	}
}
