package appref_test

import (
	"bytes"
	"context"
	"fmt"

	"rme/encoding/appref"
	"rme/encoding/liftrs"
)

func ExampleEncode() {
	// A manually supplied plan splits one byte into four 2-bit PIR records.
	// All four use the same code and row range, so query generation can be shared.
	plan := appref.Plan{Rows: 1, RecordBits: 8}
	for column := uint64(0); column < 8; column += 2 {
		plan.Databases = append(plan.Databases, appref.Partition{
			Coordinates: appref.Coordinates{RowStart: 0, ColumnStart: column, RowEnd: 1, ColumnEnd: column + 2},
			Code:        liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 2},
			Group:       0,
		})
	}
	client, stored, err := appref.Encode(plan, [][]byte{{0xa5}})
	if err != nil {
		panic(err)
	}
	storage := make(map[appref.Coordinates]*bytes.Reader)
	for _, database := range stored {
		storage[database.Coordinates] = bytes.NewReader(database.Data)
	}
	read := func(ctx context.Context, id appref.Coordinates, p []byte, offset int64) (int, error) {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		return storage[id].ReadAt(p, offset)
	}
	got, err := client.Retrieve(context.Background(), 0, read)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%x\n", got)
	// Output: a5
}
