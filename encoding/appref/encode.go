package appref

import (
	"fmt"

	"rme/encoding/cryptoref"
)

// Encode validates and executes a supplied plan. Each input record occupies
// ceil(RecordBits/8) bytes, with zero unused high bits in its last byte. It pads
// submatrix rows/columns, converts bits to field elements, and invokes both
// reference encoders. Returned client secrets and stored bytes are separate.
// Every call generates fresh permutation and per-database masking keys.
func Encode(plan Plan, records [][]byte) (*Client, []StoredDatabase, error) {
	c, err := validate(plan)
	if err != nil {
		return nil, nil, err
	}
	if uint64(len(records)) != plan.Rows {
		return nil, nil, fmt.Errorf("input has %d rows, want %d", len(records), plan.Rows)
	}
	width := int((plan.RecordBits + 7) / 8)
	for i, row := range records {
		if len(row) != width {
			return nil, nil, fmt.Errorf("application row %d has %d bytes, want %d", i, len(row), width)
		}
		if tail := plan.RecordBits % 8; tail != 0 && row[width-1]>>uint(tail) != 0 {
			return nil, nil, fmt.Errorf("application row %d has nonzero unused high bits", i)
		}
	}
	for i := range c.groups {
		key, err := cryptoref.GenerateKey()
		if err != nil {
			return nil, nil, err
		}
		permutation, err := cryptoref.NewPermutation(c.groups[i].code.Length(), key)
		if err != nil {
			return nil, nil, err
		}
		c.groups[i].permutation = permutation
	}
	stored := make([]StoredDatabase, len(c.databases))
	for i := range c.databases {
		d := &c.databases[i]
		p := d.partition
		message := make([]uint32, d.code.Dimension())
		for row := p.RowStart; row < p.RowEnd; row++ {
			for col := p.ColumnStart; col < p.ColumnEnd; col++ {
				bit := (records[int(row)][int(col/8)] >> uint(col%8)) & 1
				message[int(row-p.RowStart)] |= uint32(bit) << uint(col-p.ColumnStart)
			}
		}
		word, err := d.code.Encode(message)
		if err != nil {
			return nil, nil, err
		}
		key, err := cryptoref.GenerateKey()
		if err != nil {
			return nil, nil, err
		}
		d.masker, err = cryptoref.NewMasker(p.Code.FieldDegree, d.code.Length(), key)
		if err != nil {
			return nil, nil, err
		}
		values, err := d.masker.Encode(word, c.groups[int(p.Group)].permutation)
		if err != nil {
			return nil, nil, err
		}
		stored[i] = StoredDatabase{Coordinates: p.Coordinates, Data: pack(values, p.Code.FieldDegree)}
	}
	return c, stored, nil
}

// pack follows the LSB-first field-element bitstream contract, without padding
// between elements. A newly allocated slice leaves the final unused bits zero.
func pack(values []uint32, width int) []byte {
	out := make([]byte, (len(values)*width+7)/8)
	for i, value := range values {
		for b := 0; b < width; b++ {
			bit := i*width + b
			out[bit/8] |= byte((value>>uint(b))&1) << uint(bit%8)
		}
	}
	return out
}
