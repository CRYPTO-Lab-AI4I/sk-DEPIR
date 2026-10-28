// Package appref is a bounded application-to-PIR correctness reference. It
// executes supplied rectangular partitions using the small Lift-RS and crypto
// references. Its in-memory objects and API do not prescribe a large-scale
// implementation, storage strategy, or automatic parameter planner.
package appref

import (
	"fmt"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/cryptoref"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-go"
)

// These bounds keep both application coverage checks and dense code fixtures
// small. They are test-backend limits, not capacities of the PIR construction.
const (
	MaxRows        = 256
	MaxRecordBits  = 256
	MaxDatabases   = 64
	DummyPositions = 128
)

// Coordinates identify a nonempty half-open rectangle in the application matrix.
// They are public, deterministic identifiers scoped to one application database.
type Coordinates struct {
	RowStart    uint64
	ColumnStart uint64
	RowEnd      uint64
	ColumnEnd   uint64
}

// Partition assigns a rectangle to a reference code and a fixed query-sharing
// group. Group numbers must start at zero in first-occurrence order. Databases
// within a group must have identical code specs and compatible original rows.
type Partition struct {
	Coordinates
	Code  liftrs.Spec
	Group uint64
}

// Plan specifies the application dimensions and every PIR rectangle. Databases
// must be in lexicographic (row_start,column_start,row_end,column_end) order and
// cover the matrix exactly once. RecordBits must be at least 4, the largest
// record size in this fixture backend; this is not the target s=8..20 catalogue.
type Plan struct {
	Rows       uint64
	RecordBits uint64
	Databases  []Partition
}

// StoredDatabase is one packed, permuted, masked database. Data is caller-owned
// and contains no client keys. A storage reader identifies it using Coordinates.
type StoredDatabase struct {
	Coordinates
	Data []byte
}

type database struct {
	partition Partition
	code      *liftrs.Code
	masker    *cryptoref.Masker
}

type group struct {
	members     []int
	code        *liftrs.Code
	permutation *cryptoref.Permutation
}

// Client retains the private encoding/query state. Obtain it through Encode.
// It is immutable after construction and supports concurrent Retrieve calls
// with independently safe storage readers. It does not retain application data.
type Client struct {
	plan      Plan
	databases []database
	groups    []group
}

// Plan returns a copy of the public layout, without secret keys or permutations.
// This reference exposes layout metadata in memory; durable manifest persistence
// and reopening with private state are separate work.
func (c *Client) Plan() Plan {
	p := c.plan
	p.Databases = append([]Partition(nil), p.Databases...)
	return p
}

// validate constructs reusable code data once per distinct spec. It does not
// alter the selected partition or replace an unsupported code with another one.
func validate(plan Plan) (*Client, error) {
	if plan.Rows == 0 || plan.Rows > MaxRows || plan.RecordBits < 4 || plan.RecordBits > MaxRecordBits {
		return nil, fmt.Errorf("reference application dimensions require rows in [1,%d] and record bits in [4,%d]", MaxRows, MaxRecordBits)
	}
	if len(plan.Databases) == 0 || len(plan.Databases) > MaxDatabases {
		return nil, fmt.Errorf("reference plan requires 1..%d databases", MaxDatabases)
	}
	c := &Client{plan: plan}
	c.plan.Databases = append([]Partition(nil), plan.Databases...)
	coverage := make([]bool, int(plan.Rows*plan.RecordBits))
	codes := make(map[liftrs.Spec]*liftrs.Code)
	for i, p := range c.plan.Databases {
		if p.RowStart >= p.RowEnd || p.RowEnd > plan.Rows || p.ColumnStart >= p.ColumnEnd || p.ColumnEnd > plan.RecordBits {
			return nil, fmt.Errorf("database %d has invalid rectangle %+v", i, p.Coordinates)
		}
		if i > 0 && !less(c.plan.Databases[i-1].Coordinates, p.Coordinates) {
			return nil, fmt.Errorf("databases must have unique coordinates in canonical order")
		}
		code := codes[p.Code]
		if code == nil {
			var err error
			code, err = liftrs.Preprocess(p.Code)
			if err != nil {
				return nil, fmt.Errorf("database %d code: %w", i, err)
			}
			codes[p.Code] = code
		}
		if p.RowEnd-p.RowStart > uint64(code.Dimension()) || p.ColumnEnd-p.ColumnStart > uint64(p.Code.FieldDegree) {
			return nil, fmt.Errorf("database %d rectangle exceeds verified code capacity", i)
		}
		if p.Group > uint64(len(c.groups)) {
			return nil, fmt.Errorf("query-sharing groups must be numbered in first-occurrence order")
		}
		if p.Group == uint64(len(c.groups)) {
			c.groups = append(c.groups, group{code: code})
		}
		g := &c.groups[int(p.Group)]
		if g.code.Spec() != p.Code {
			return nil, fmt.Errorf("query-sharing group %d has incompatible code configurations", p.Group)
		}
		for _, member := range g.members {
			other := c.databases[member].partition
			overlap := p.RowStart < other.RowEnd && other.RowStart < p.RowEnd
			if overlap && p.RowStart != other.RowStart {
				return nil, fmt.Errorf("query-sharing group %d has incompatible row ranges", p.Group)
			}
		}
		for row := p.RowStart; row < p.RowEnd; row++ {
			for col := p.ColumnStart; col < p.ColumnEnd; col++ {
				index := int(row*plan.RecordBits + col)
				if coverage[index] {
					return nil, fmt.Errorf("overlapping rectangles at row %d column %d", row, col)
				}
				coverage[index] = true
			}
		}
		c.databases = append(c.databases, database{partition: p, code: code})
		g.members = append(g.members, i)
	}
	for _, covered := range coverage {
		if !covered {
			return nil, fmt.Errorf("partition leaves application bits uncovered")
		}
	}
	return c, nil
}

func less(a, b Coordinates) bool {
	x := [4]uint64{a.RowStart, a.ColumnStart, a.RowEnd, a.ColumnEnd}
	y := [4]uint64{b.RowStart, b.ColumnStart, b.RowEnd, b.ColumnEnd}
	for i := range x {
		if x[i] != y[i] {
			return x[i] < y[i]
		}
	}
	return false
}
