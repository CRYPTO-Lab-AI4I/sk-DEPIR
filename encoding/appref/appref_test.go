package appref

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"sync"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
)

func fixturePlan() Plan {
	small := liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 2}
	medium := liftrs.Spec{FieldDegree: 3, Dimension: 2, CurveDegree: 2}
	large := liftrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 2}
	return Plan{Rows: 8, RecordBits: 11, Databases: []Partition{
		{Coordinates{0, 0, 5, 2}, small, 0},
		{Coordinates{0, 2, 8, 5}, medium, 1},
		{Coordinates{0, 5, 4, 8}, medium, 1},
		{Coordinates{0, 8, 8, 11}, large, 2},
		{Coordinates{4, 5, 8, 8}, medium, 3},
		{Coordinates{5, 0, 8, 2}, small, 0},
	}}
}

func fixtureRecords() [][]byte {
	out := make([][]byte, 8)
	for i := range out {
		out[i] = []byte{byte(17*i + 0x36), byte(i & 7)}
	}
	return out
}

func memoryReader(stored []StoredDatabase) ReadAt {
	data := make(map[Coordinates]*bytes.Reader)
	for _, db := range stored {
		data[db.Coordinates] = bytes.NewReader(db.Data)
	}
	return func(ctx context.Context, id Coordinates, p []byte, offset int64) (int, error) {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		r := data[id]
		if r == nil {
			return 0, errors.New("unknown database")
		}
		return r.ReadAt(p, offset)
	}
}

func TestGeneralPartitionRecoveryAndParticipation(t *testing.T) {
	plan := fixturePlan()
	records := fixtureRecords()
	c, stored, err := Encode(plan, records)
	if err != nil {
		t.Fatal(err)
	}
	baseRead := memoryReader(stored)
	type access struct {
		offset int64
		length int
	}
	for row := uint64(0); row < plan.Rows; row++ {
		trace := make(map[Coordinates][]access)
		read := func(ctx context.Context, id Coordinates, p []byte, offset int64) (int, error) {
			trace[id] = append(trace[id], access{offset, len(p)})
			return baseRead(ctx, id, p, offset)
		}
		got, err := c.Retrieve(context.Background(), row, read)
		if err != nil || !bytes.Equal(got, records[row]) {
			t.Fatalf("row %d: got %x want %x, error %v", row, got, records[row], err)
		}
		for _, g := range c.groups {
			var first []access
			for _, member := range g.members {
				d := c.databases[member]
				calls := trace[d.partition.Coordinates]
				want := (1 << uint(d.partition.Code.FieldDegree)) - 1 + DummyPositions
				if len(calls) != want {
					t.Fatalf("row %d database %+v: %d calls want %d", row, d.partition.Coordinates, len(calls), want)
				}
				if first != nil && !reflect.DeepEqual(first, calls) {
					t.Fatal("query-generation sharing changed the submitted read list")
				}
				first = calls
			}
		}
		if len(trace) != len(plan.Databases) {
			t.Fatal("skipped irrelevant database")
		}
	}
	if c.databases[0].code != c.databases[5].code || c.databases[1].code != c.databases[4].code {
		t.Fatal("code preprocessing was not reused across matching specs")
	}
}

func TestPaddingAndBitConversionGoldenParts(t *testing.T) {
	plan := fixturePlan()
	records := make([][]byte, 8)
	for i := range records {
		records[i] = []byte{0xd6, 0x05}
	}
	c, stored, err := Encode(plan, records)
	if err != nil {
		t.Fatal(err)
	}
	read := memoryReader(stored)
	// LSB-first slices of d6 05 at [0,2),[2,5),[5,8),[8,11).
	wantPart := map[uint64]uint32{0: 2, 2: 5, 5: 6, 8: 5}
	for i, d := range c.databases {
		positions := make([]int, d.code.Length())
		values := make([]uint32, len(positions))
		for pos := range positions {
			positions[pos] = pos
			values[pos], err = readElement(context.Background(), read, d.partition.Coordinates, pos, d.partition.Code.FieldDegree)
			if err != nil {
				t.Fatal(err)
			}
		}
		plain, err := d.masker.Unmask(positions, values)
		if err != nil {
			t.Fatal(err)
		}
		for j, codePosition := range d.code.SystematicPositions() {
			want := uint32(0)
			if j < int(d.partition.RowEnd-d.partition.RowStart) {
				want = wantPart[d.partition.ColumnStart]
			}
			storage, err := c.groups[int(d.partition.Group)].permutation.Map([]int{codePosition})
			if err != nil {
				t.Fatal(err)
			}
			if plain[storage[0]] != want {
				t.Fatalf("database %d message row %d: got %d want %d", i, j, plain[storage[0]], want)
			}
		}
		if len(stored[i].Data) != (d.code.Length()*d.partition.Code.FieldDegree+7)/8 {
			t.Fatal("wrong packed length")
		}
	}
}

func TestAllApplicationByteValues(t *testing.T) {
	plan := Plan{Rows: 1, RecordBits: 8}
	for col := uint64(0); col < 8; col += 2 {
		plan.Databases = append(plan.Databases, Partition{Coordinates{0, col, 1, col + 2}, liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 2}, 0})
	}
	for value := 0; value < 256; value++ {
		record := []byte{byte(value)}
		c, stored, err := Encode(plan, [][]byte{record})
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.Retrieve(context.Background(), 0, memoryReader(stored))
		if err != nil || !bytes.Equal(got, record) {
			t.Fatalf("byte %d: got %x error %v", value, got, err)
		}
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func TestRepeatedCurveEntriesAndDeterministicFixtures(t *testing.T) {
	c, stored, err := Encode(fixturePlan(), fixtureRecords())
	if err != nil {
		t.Fatal(err)
	}
	for row := uint64(0); row < c.plan.Rows; row++ {
		got, err := c.retrieve(context.Background(), row, memoryReader(stored), zeroReader{})
		if err != nil || !bytes.Equal(got, fixtureRecords()[row]) {
			t.Fatal("repeated constant-curve recovery failed", err)
		}
	}
	q, err := makeQuery(c.groups[0], 0, zeroReader{})
	if err != nil {
		t.Fatal(err)
	}
	if len(q.real) != 3 || len(q.positions) != 3+DummyPositions {
		t.Fatal("wrong real/dummy shape")
	}
	for _, index := range q.real {
		if q.positions[index] != q.positions[q.real[0]] {
			t.Fatal("fixture did not force repeated curve point")
		}
	}
	again, err := makeQuery(c.groups[0], 0, zeroReader{})
	if err != nil || !reflect.DeepEqual(q, again) {
		t.Fatal("test entropy did not give reproducible query", err)
	}
}

func TestPlanAndInputValidation(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Plan)
	}{
		{"zero rows", func(p *Plan) { p.Rows = 0 }},
		{"too many rows", func(p *Plan) { p.Rows = MaxRows + 1 }},
		{"small records", func(p *Plan) { p.RecordBits = 3 }},
		{"large records", func(p *Plan) { p.RecordBits = MaxRecordBits + 1 }},
		{"empty plan", func(p *Plan) { p.Databases = nil }},
		{"too many databases", func(p *Plan) { p.Databases = make([]Partition, MaxDatabases+1) }},
		{"invalid rows", func(p *Plan) { p.Databases[0].RowEnd = 9 }},
		{"invalid columns", func(p *Plan) { p.Databases[0].ColumnStart = 2 }},
		{"unsupported code", func(p *Plan) { p.Databases[0].Code.FieldDegree = 8 }},
		{"capacity", func(p *Plan) { p.Databases[0].RowEnd = 6 }},
		{"width", func(p *Plan) { p.Databases[0].ColumnEnd = 3 }},
		{"gap", func(p *Plan) { p.Databases = p.Databases[:5] }},
		{"overlap", func(p *Plan) { p.Databases[4].RowStart = 3; p.Databases[4].Group = 3 }},
		{"order", func(p *Plan) { p.Databases[0], p.Databases[1] = p.Databases[1], p.Databases[0] }},
		{"group numbering", func(p *Plan) { p.Databases[0].Group = 2 }},
		{"group spec", func(p *Plan) { p.Databases[2].Code.FieldDegree = 4 }},
		{"overlapping group rows", func(p *Plan) { p.Databases[4].Group = 1 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := fixturePlan()
			tc.change(&p)
			if c, data, err := Encode(p, fixtureRecords()); err == nil || c != nil || data != nil {
				t.Fatal("accepted invalid plan or returned partial output")
			}
		})
	}
	for _, records := range [][][]byte{nil, {{0}}, {{0, 0}, {0, 0}}} {
		if _, _, err := Encode(fixturePlan(), records); err == nil {
			t.Fatal("accepted malformed rows")
		}
	}
	records := fixtureRecords()
	records[0][1] = 0x80
	if _, _, err := Encode(fixturePlan(), records); err == nil {
		t.Fatal("accepted nonzero bits outside application width")
	}
	records = fixtureRecords()
	records[2] = nil
	if _, _, err := Encode(fixturePlan(), records); err == nil {
		t.Fatal("accepted wrong row byte width")
	}
}

func TestReadFailuresAndCancellation(t *testing.T) {
	c, stored, err := Encode(fixturePlan(), fixtureRecords())
	if err != nil {
		t.Fatal(err)
	}
	reader := memoryReader(stored)
	failure := errors.New("storage failure")
	cases := []struct {
		read ReadAt
		want error
	}{
		{func(context.Context, Coordinates, []byte, int64) (int, error) { return 0, failure }, failure},
		{func(context.Context, Coordinates, []byte, int64) (int, error) { return 0, nil }, io.ErrUnexpectedEOF},
		{func(context.Context, Coordinates, []byte, int64) (int, error) { return 0, io.EOF }, io.EOF},
	}
	for _, tc := range cases {
		if got, err := c.Retrieve(context.Background(), 0, tc.read); got != nil || !errors.Is(err, tc.want) {
			t.Fatalf("read failure: got %x error %v want %v", got, err, tc.want)
		}
	}
	for _, n := range []int{-1, 99} {
		read := func(context.Context, Coordinates, []byte, int64) (int, error) { return n, nil }
		if got, err := c.Retrieve(context.Background(), 0, read); got != nil || err == nil {
			t.Fatal("accepted invalid byte count")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := c.Retrieve(ctx, 0, reader); got != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("ignored canceled context", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	read := func(ctx context.Context, id Coordinates, p []byte, offset int64) (int, error) {
		n, err := reader(ctx, id, p, offset)
		cancel()
		return n, err
	}
	if got, err := c.Retrieve(ctx, 0, read); got != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("returned partial record after cancellation", err)
	}
	if got, err := c.Retrieve(context.Background(), 8, reader); got != nil || err == nil {
		t.Fatal("accepted invalid row")
	}
	if _, err := c.Retrieve(nil, 0, reader); err == nil {
		t.Fatal("accepted nil context")
	}
	if _, err := c.Retrieve(context.Background(), 0, nil); err == nil {
		t.Fatal("accepted nil reader")
	}
	var empty Client
	if _, err := empty.Retrieve(context.Background(), 0, reader); err == nil {
		t.Fatal("accepted uninitialized client")
	}
	var absent *Client
	if _, err := absent.Retrieve(context.Background(), 0, reader); err == nil {
		t.Fatal("accepted nil client")
	}
	calls := 0
	countRead := func(ctx context.Context, id Coordinates, p []byte, offset int64) (int, error) {
		calls++
		return reader(ctx, id, p, offset)
	}
	if got, err := c.retrieve(context.Background(), 0, countRead, errorReader{failure}); got != nil || !errors.Is(err, failure) || calls != 0 {
		t.Fatal("query entropy failure was not propagated before reads", err)
	}
}

func TestOwnershipReencodingAndConcurrency(t *testing.T) {
	plan := fixturePlan()
	records := fixtureRecords()
	c, stored, err := Encode(plan, records)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), records[0]...)
	_, fresh, err := Encode(plan, records)
	if err != nil {
		t.Fatal(err)
	}
	// Compare a whole GF(16) ciphertext fixture, never individual tiny masks.
	if bytes.Equal(stored[3].Data, fresh[3].Data) {
		t.Fatal("re-encoding unexpectedly reused its full encrypted fixture")
	}
	if !bytes.Equal(before, records[0]) {
		t.Fatal("Encode modified input")
	}
	plan.Databases[0].RowEnd = 1
	copyPlan := c.Plan()
	copyPlan.Databases[0].RowEnd = 2
	if c.Plan().Databases[0].RowEnd != 5 {
		t.Fatal("caller can mutate retained plan")
	}
	reader := memoryReader(stored)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(row uint64) {
			defer wg.Done()
			got, err := c.Retrieve(context.Background(), row, reader)
			if err != nil || !bytes.Equal(got, records[row]) {
				t.Errorf("concurrent row %d: got %x error %v", row, got, err)
			}
		}(uint64(i))
	}
	wg.Wait()
}

func TestPackedGoldenVector(t *testing.T) {
	if got := pack([]uint32{1, 1}, 9); !bytes.Equal(got, []byte{1, 2, 0}) {
		t.Fatalf("9-bit golden packing = %x", got)
	}
	// The middle 3-bit element straddles a byte, and a complete read may return EOF.
	data := pack([]uint32{0, 1, 2, 3, 4, 5, 6, 7}, 3)
	read := func(_ context.Context, _ Coordinates, p []byte, offset int64) (int, error) {
		n, err := bytes.NewReader(data).ReadAt(p, offset)
		if n == len(p) {
			return n, io.EOF
		}
		return n, err
	}
	for i := 0; i < 8; i++ {
		got, err := readElement(context.Background(), read, Coordinates{}, i, 3)
		if err != nil || got != uint32(i) {
			t.Fatalf("packed element %d: got %d error %v", i, got, err)
		}
	}
}

func TestThreeDimensionalAndHighCurveDegreeFixtures(t *testing.T) {
	for _, fixture := range []struct {
		spec liftrs.Spec
		rows uint64
	}{
		{liftrs.Spec{FieldDegree: 2, Dimension: 3, CurveDegree: 1}, 7},
		{liftrs.Spec{FieldDegree: 2, Dimension: 3, CurveDegree: 2}, 7},
		{liftrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 14}, 1},
	} {
		s := uint64(fixture.spec.FieldDegree)
		plan := Plan{Rows: fixture.rows, RecordBits: 2 * s, Databases: []Partition{
			{Coordinates{0, 0, fixture.rows, s}, fixture.spec, 0},
			{Coordinates{0, s, fixture.rows, 2 * s}, fixture.spec, 0},
		}}
		records := make([][]byte, fixture.rows)
		for i := range records {
			records[i] = []byte{byte((3*i + 5) % int(1<<(2*s)))}
		}
		c, stored, err := Encode(plan, records)
		if err != nil {
			t.Fatal(err)
		}
		for row := range records {
			got, err := c.Retrieve(context.Background(), uint64(row), memoryReader(stored))
			if err != nil || !bytes.Equal(got, records[row]) {
				t.Fatalf("spec %+v row %d: got %x error %v", fixture.spec, row, got, err)
			}
		}
	}
}
