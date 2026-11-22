package liftrs

import "fmt"

func discoveryResultMode(value string) (string, error) {
	if value == "" {
		return "resident", nil
	}
	if value != "resident" && value != "stream" {
		return "", fmt.Errorf("discovery result mode")
	}
	return value, nil
}

// discoveryUnion consumes one complete local result in ordered tiles. The bitmap
// persists across samples; Added counts only newly discovered nonzero positions.
// A rejected tile changes neither the bitmap nor counters. Callers must abort a
// failed collective before reusing its GPU workspace or accepting sample counts.
type discoveryUnion[Word transposeWord] struct {
	Found                     exponentSet
	Words, Next, Added, Limit uint64
}

func (u *discoveryUnion[Word]) Consume(offset uint64, values []Word) error {
	if u.Words > uint64(^uint(0)>>1) || u.Limit < 2 || u.Next > u.Words || offset != u.Next || len(values) == 0 || uint64(len(values)) > u.Words-u.Next || uint64(len(u.Found)) < (u.Words+63)/64 {
		return fmt.Errorf("discovery tile extent/order/bitmap admission")
	}
	for _, value := range values {
		if uint64(value) >= u.Limit {
			return fmt.Errorf("discovery result field range")
		}
	}
	for j, value := range values {
		index := int(offset + uint64(j))
		if value != 0 && !u.Found.has(index) {
			u.Found.add(index)
			u.Added++
		}
	}
	u.Next += uint64(len(values))
	return nil
}

func (u *discoveryUnion[Word]) Complete() (uint64, error) {
	if u.Next != u.Words {
		return 0, fmt.Errorf("incomplete discovery result")
	}
	return u.Added, nil
}
